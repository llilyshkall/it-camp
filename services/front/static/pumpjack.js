/**
 * Pumpjack Animation Component
 * Адаптированная версия анимации pumpjack для встраивания в сайты
 */
class PumpjackAnimation {
  constructor(canvas, options = {}) {
    this.canvas = typeof canvas === 'string' ? document.querySelector(canvas) : canvas;
    this.ctx = this.canvas.getContext('2d');
    this.options = {
      width: 900,
      height: 600,
      speed: 0.4,
      stroke: 18,
      showPivots: false,
      autoStart: true,
      ...options
    };
    
    this.running = false;
    this.freq = this.options.speed;
    this.omega = 2.0 * Math.PI * this.freq;
    this.strokeTargetDeg = this.options.stroke;
    this.phi = 0.0;
    this.lastTs = null;
    this.lastTheta = 0.0;
    this.rodTopY = 235.0;
    
    // Geometry configuration
    this.config = {
      OX: 450.0,   // beam pivot (top of tower)
      OY: 235.0,
      CX: 740.0,   // crank center
      CY: 420.0,
      beamRight: 190.0,  // attachment point to pitman (right of O)
      beamLeft: 210.0,   // horsehead distance to left end
      crankR: 56.0,      // crank radius (auto-calibrated)
      pitmanL: 315.0,    // pitman length
      rodBottomY: 540.0, // polished rod shoe Y
      easeTau: 0.09      // easing time constant
    };
    
    this.init();
  }
  
  init() {
    this.calibrateCrankRadius(this.strokeTargetDeg);
    this.updateUI();
    
    if (this.options.autoStart) {
      this.start();
    }
  }
  
  // Kinematics helpers
  beamAttach(theta) {
    const bx = this.config.OX + this.config.beamRight * Math.cos(theta);
    const by = this.config.OY + this.config.beamRight * Math.sin(theta);
    return [bx, by];
  }
  
  leftEnd(theta) {
    const hx = this.config.OX - this.config.beamLeft * Math.cos(theta);
    const hy = this.config.OY - this.config.beamLeft * Math.sin(theta);
    return [hx, hy];
  }
  
  crankPin(phi) {
    const px = this.config.CX + this.config.crankR * Math.cos(phi);
    const py = this.config.CY + this.config.crankR * Math.sin(phi);
    return [px, py];
  }
  
  solveTheta(phi, theta0) {
    const [px, py] = this.crankPin(phi);
    let th = theta0;
    
    for (let i = 0; i < 8; i++) {
      const bx = this.config.OX + this.config.beamRight * Math.cos(th);
      const by = this.config.OY + this.config.beamRight * Math.sin(th);
      const dx = bx - px;
      const dy = by - py;
      const f = dx * dx + dy * dy - this.config.pitmanL * this.config.pitmanL;
      
      if (Math.abs(f) < 1e-7) break;
      
      // derivative: 2*(B - P) • dB/dθ ;  dB/dθ = (-br*sinθ, br*cosθ)
      const dBx = -this.config.beamRight * Math.sin(th);
      const dBy = this.config.beamRight * Math.cos(th);
      const fp = 2.0 * (dx * dBx + dy * dBy);
      
      // avoid zero derivative
      if (Math.abs(fp) < 1e-9) {
        fp = fp >= 0 ? 1e-9 : -1e-9;
      }
      
      th -= f / fp;
    }
    
    return th;
  }
  
  computeThetaRange() {
    let thMin = 1e9;
    let thMax = -1e9;
    let th = 0.0;
    
    // sample uniformly
    for (let i = 0; i < 200; i++) {
      const ph = 2 * Math.PI * (i / 200.0);
      th = this.solveTheta(ph, th);
      if (th < thMin) thMin = th;
      if (th > thMax) thMax = th;
    }
    
    return [thMin, thMax];
  }
  
  calibrateCrankRadius(targetDeg) {
    const target = targetDeg * Math.PI / 180.0;
    // search bounds (pixels)
    let lo = 24.0, hi = 110.0;
    let bestR = this.config.crankR, bestErr = 1e9;
    
    for (let i = 0; i < 22; i++) {
      const mid = 0.5 * (lo + hi);
      this.config.crankR = mid;
      const [tmin, tmax] = this.computeThetaRange();
      const amp = 0.5 * (tmax - tmin);
      const err = Math.abs(amp - target);
      
      if (err < bestErr) {
        bestErr = err;
        bestR = mid;
      }
      
      if (amp < target) {
        lo = mid;
      } else {
        hi = mid;
      }
    }
    
    this.config.crankR = bestR;
    // after changing geometry, update last_theta using current phi
    this.lastTheta = this.solveTheta(this.phi, this.lastTheta);
  }
  
  drawFrame(theta, px, py, bx, by, dt) {
    // Clear canvas
    this.ctx.clearRect(0, 0, this.canvas.width, this.canvas.height);
    
    // Polished rod: eased vertical motion
    const [hx, hy] = this.leftEnd(theta);
    // easing factor alpha in [0..1]
    const s = Math.min(1.0, dt / Math.max(1e-6, this.config.easeTau));
    const alpha = 0.5 * (1.0 - Math.cos(Math.PI * s));  // cosine smooth
    this.rodTopY = this.rodTopY + (hy - this.rodTopY) * alpha;
    
    // Draw GROUND
    this.ctx.fillStyle = 'rgba(10, 115, 183, 0.15)';
    this.ctx.fillRect(0, 540, 900, 60);
    
    // Draw TOWER
    this.ctx.fillStyle = '#0a73b7';
    this.ctx.fillRect(350, 520, 200, 20);
    this.ctx.beginPath();
    this.ctx.moveTo(380, 520);
    this.ctx.lineTo(520, 520);
    this.ctx.lineTo(480, 260);
    this.ctx.lineTo(420, 260);
    this.ctx.closePath();
    this.ctx.fill();
    
    // Draw braces
    this.ctx.strokeStyle = '#0a73b7';
    this.ctx.lineWidth = 6;
    this.ctx.beginPath();
    this.ctx.moveTo(398, 520);
    this.ctx.lineTo(438, 260);
    this.ctx.moveTo(502, 520);
    this.ctx.lineTo(462, 260);
    this.ctx.moveTo(380, 520);
    this.ctx.lineTo(520, 520);
    this.ctx.moveTo(420, 260);
    this.ctx.lineTo(438, 260);
    this.ctx.moveTo(462, 260);
    this.ctx.lineTo(480, 260);
    this.ctx.stroke();
    
    // Draw BEAM + HORSEHEAD (rotates around pivot Ox,Oy)
    this.ctx.save();
    this.ctx.translate(this.config.OX, this.config.OY);
    this.ctx.rotate(theta);
    this.ctx.translate(-this.config.OX, -this.config.OY);
    
    // Main walking beam body
    this.ctx.fillStyle = '#0a73b7';
    this.ctx.fillRect(280, 220, 360, 30);
    this.ctx.fillStyle = 'rgba(10, 115, 183, 0.75)';
    this.ctx.fillRect(430, 220, 60, 30);
    this.ctx.fillRect(500, 220, 40, 30);
    this.ctx.fillStyle = '#0a73b7';
    
    // Horsehead at left end
    this.ctx.beginPath();
    this.ctx.moveTo(280, 221);
    this.ctx.bezierCurveTo(260, 210, 230, 205, 210, 230);
    this.ctx.lineTo(210, 240);
    this.ctx.bezierCurveTo(235, 270, 270, 255, 280, 250);
    this.ctx.closePath();
    this.ctx.fill();
    
    // Clamp for polished rod
    this.ctx.fillRect(265, 232, 14, 8);
    this.ctx.restore();
    
    // Draw CRANK (disc + arm)
    this.ctx.save();
    this.ctx.translate(this.config.CX, this.config.CY);
    this.ctx.rotate(this.phi);
    this.ctx.translate(-this.config.CX, -this.config.CY);

    this.ctx.fillStyle = '#0a73b7';
    this.ctx.beginPath();
    this.ctx.arc(this.config.CX, this.config.CY, 32, 0, 2 * Math.PI);
    this.ctx.fill();
    
    this.ctx.restore();
    
    // Draw PITMAN (connecting rod) - крепится к математической точке
    this.ctx.strokeStyle = '#0a73b7';
    this.ctx.lineWidth = 6;
    this.ctx.beginPath();
    // Шатун крепится к математической точке кривошипа
    this.ctx.moveTo(px, py);
    this.ctx.lineTo(bx, by);
    this.ctx.stroke();
    
    // Draw POLISHED ROD (vertical) + shoe
    this.ctx.strokeStyle = '#0a73b7';
    this.ctx.beginPath();
    this.ctx.moveTo(hx, this.rodTopY);
    this.ctx.lineTo(hx, this.config.rodBottomY);
    this.ctx.stroke();
    
    // Draw shoe
    this.ctx.fillStyle = '#0a73b7';
    this.ctx.beginPath();
    this.ctx.moveTo(225, 540);  // Расширил левый край (было 235)
    this.ctx.lineTo(295, 540);  // Расширил правый край (было 285)
    this.ctx.lineTo(260, 560);  // Вершина остается по центру
    this.ctx.closePath();
    this.ctx.fill();
    
    // Draw optional pivot markers
    if (this.options.showPivots) {
      this.ctx.fillStyle = '#0a73b7';
      this.ctx.beginPath();
      this.ctx.arc(this.config.OX, this.config.OY, 6, 0, 2 * Math.PI);
      this.ctx.fill();
      this.ctx.beginPath();
      this.ctx.arc(this.config.CX, this.config.CY, 6, 0, 2 * Math.PI);
      this.ctx.fill();
      this.ctx.beginPath();
      this.ctx.arc(bx, by, 6, 0, 2 * Math.PI);
      this.ctx.fill();
      this.ctx.beginPath();
      this.ctx.arc(px, py, 6, 0, 2 * Math.PI);
      this.ctx.fill();
      this.ctx.fillStyle = '#0a73b7';
    }
  }
  
  // Animation loop
  frame(ts) {
    if (this.lastTs === null) {
      this.lastTs = ts;
    }
    
    const dt = (ts - this.lastTs) / 1000.0;
    this.lastTs = ts;
    
    if (this.running) {
      this.phi += this.omega * dt;
    }
    
    // Solve for beam theta given crank angle
    this.lastTheta = this.solveTheta(this.phi, this.lastTheta);
    const [bx, by] = this.beamAttach(this.lastTheta);
    const [px, py] = this.crankPin(this.phi);
    
    this.drawFrame(this.lastTheta, px, py, bx, by, dt);
    
    if (this.running) {
      this.animationId = requestAnimationFrame((ts) => this.frame(ts));
    }
  }
  
  // Public API methods
  start() {
    if (!this.running) {
      this.running = true;
      this.animationId = requestAnimationFrame((ts) => this.frame(ts));
    }
  }
  
  pause() {
    this.running = false;
    if (this.animationId) {
      cancelAnimationFrame(this.animationId);
    }
  }
  
  reset() {
    this.phi = 0.0;
    this.lastTs = null;
    this.lastTheta = this.solveTheta(this.phi, this.lastTheta);
    const [, hy] = this.leftEnd(this.lastTheta);
    this.rodTopY = hy;
    
    // Update display immediately
    const [bx, by] = this.beamAttach(this.lastTheta);
    const [px, py] = this.crankPin(this.phi);
    this.drawFrame(this.lastTheta, px, py, bx, by, 0);
  }
  
  setSpeed(speed) {
    this.freq = speed;
    this.omega = 2.0 * Math.PI * this.freq;
  }
  
  setStroke(stroke) {
    this.strokeTargetDeg = stroke;
    this.calibrateCrankRadius(stroke);
  }
  
  togglePivots(show) {
    this.options.showPivots = show;
  }
  
  updateUI() {
    // Canvas-based UI doesn't need DOM updates
  }
  
  // Cleanup method
  destroy() {
    if (this.animationId) {
      cancelAnimationFrame(this.animationId);
    }
  }
  
  // Static method to create multiple instances
  static createMultiple(canvases, options = {}) {
    return canvases.map(canvas => new PumpjackAnimation(canvas, options));
  }
}

// Export for different module systems
if (typeof module !== 'undefined' && module.exports) {
  module.exports = PumpjackAnimation;
} else if (typeof define === 'function' && define.amd) {
  define([], function() { return PumpjackAnimation; });
} else {
  window.PumpjackAnimation = PumpjackAnimation;
}
