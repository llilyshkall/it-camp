// ===== ДОПОЛНИТЕЛЬНЫЕ ФУНКЦИИ ДЛЯ МОБИЛЬНЫХ УСТРОЙСТВ =====

(function() {
  'use strict';

  // Определяем, является ли устройство мобильным
  const isMobile = () => {
    return window.innerWidth <= 768 || 
           /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(navigator.userAgent);
  };

  // Определяем, является ли устройство планшетом
  const isTablet = () => {
    return window.innerWidth > 768 && window.innerWidth <= 1024;
  };

  // Определяем, является ли устройство телефоном
  const isPhone = () => {
    return window.innerWidth <= 480;
  };

  // Улучшенная обработка касаний для мобильных устройств
  const enhanceTouchExperience = () => {
    if (!isMobile()) return;

    // Добавляем haptic feedback для кнопок (если поддерживается)
    const addHapticFeedback = (element) => {
      element.addEventListener('touchstart', () => {
        if (navigator.vibrate) {
          navigator.vibrate(10);
        }
      });
    };

    // Применяем haptic feedback ко всем кнопкам
    document.querySelectorAll('.btn, .nav-menu-link, .custom-select-option').forEach(addHapticFeedback);

    // Улучшенная обработка свайпов для навигации
    let startX = 0;
    let startY = 0;
    let currentX = 0;
    let currentY = 0;

    const handleTouchStart = (e) => {
      startX = e.touches[0].clientX;
      startY = e.touches[0].clientY;
    };

    const handleTouchMove = (e) => {
      currentX = e.touches[0].clientX;
      currentY = e.touches[0].clientY;
    };

    const handleTouchEnd = () => {
      const diffX = startX - currentX;
      const diffY = startY - currentY;
      const minSwipeDistance = 50;

      if (Math.abs(diffX) > Math.abs(diffY) && Math.abs(diffX) > minSwipeDistance) {
        // Горизонтальный свайп
        if (diffX > 0) {
          // Свайп влево - закрыть меню навигации
          const navDropdown = document.getElementById('nav-dropdown');
          if (navDropdown && !navDropdown.hidden) {
            closeNavMenu();
          }
        } else {
          // Свайп вправо - открыть меню навигации
          const navDropdown = document.getElementById('nav-dropdown');
          if (navDropdown && navDropdown.hidden) {
            openNavMenu();
          }
        }
      }
    };

    // Добавляем обработчики свайпов к body
    document.body.addEventListener('touchstart', handleTouchStart, { passive: true });
    document.body.addEventListener('touchmove', handleTouchMove, { passive: true });
    document.body.addEventListener('touchend', handleTouchEnd, { passive: true });
  };

  // Улучшенная навигация для мобильных устройств
  const enhanceMobileNavigation = () => {
    if (!isMobile()) return;

    const navToggle = document.getElementById('nav-toggle');
    const navDropdown = document.getElementById('nav-dropdown');
    const navMenu = document.getElementById('nav-menu');

    if (!navToggle || !navDropdown || !navMenu) return;

    // Функция открытия меню
    const openNavMenu = () => {
      navDropdown.hidden = false;
      navToggle.setAttribute('aria-expanded', 'true');
      
      // Добавляем класс для анимации
      requestAnimationFrame(() => {
        navDropdown.classList.add('open');
        document.body.style.overflow = 'hidden'; // Блокируем скролл
      });
      
      // Фокус на первый элемент меню для доступности
      setTimeout(() => {
        const firstLink = navMenu.querySelector('.nav-menu-link');
        if (firstLink) firstLink.focus();
      }, 100);
    };

    // Функция закрытия меню
    const closeNavMenu = () => {
      navDropdown.classList.remove('open');
      navToggle.setAttribute('aria-expanded', 'false');
      document.body.style.overflow = ''; // Восстанавливаем скролл
      
      // Ждем окончания анимации перед скрытием
      setTimeout(() => {
        navDropdown.hidden = true;
      }, 300);
      
      navToggle.focus();
    };

    // Обработчик клика по кнопке навигации
    navToggle.addEventListener('click', (e) => {
      e.preventDefault();
      if (navDropdown.hidden) {
        openNavMenu();
      } else {
        closeNavMenu();
      }
    });

    // Закрытие при клике вне меню
    document.addEventListener('click', (e) => {
      if (!navToggle.contains(e.target) && !navDropdown.contains(e.target)) {
        closeNavMenu();
      }
    });

    // Закрытие по Escape
    document.addEventListener('keydown', (e) => {
      if (e.key === 'Escape' && !navDropdown.hidden) {
        closeNavMenu();
      }
    });

    // Закрытие при скролле
    let scrollTimeout;
    document.addEventListener('scroll', () => {
      if (!navDropdown.hidden) {
        clearTimeout(scrollTimeout);
        scrollTimeout = setTimeout(() => {
          closeNavMenu();
        }, 100);
      }
    }, { passive: true });

    // Закрытие при изменении размера окна
    window.addEventListener('resize', () => {
      if (!navDropdown.hidden) {
        closeNavMenu();
      }
    });

    // Обработка кликов по ссылкам в меню
    navMenu.addEventListener('click', (e) => {
      const link = e.target.closest('.nav-menu-link');
      if (link) {
        const target = link.getAttribute('data-target');
        if (target) {
          e.preventDefault();
          closeNavMenu();
          // Плавный скролл к цели
          const targetElement = document.querySelector(target);
          if (targetElement) {
            targetElement.scrollIntoView({ behavior: 'smooth', block: 'start' });
          }
        }
      }
    });
  };

  // Улучшенная работа с фильтрами на мобильных
  const enhanceMobileFilters = () => {
    if (!isMobile()) return;

    const filterToggles = document.querySelectorAll('.custom-select-toggle');
    
    filterToggles.forEach(toggle => {
      const wrapper = toggle.closest('.custom-select');
      const menu = wrapper.querySelector('.custom-select-menu');
      const options = menu.querySelectorAll('.custom-select-option');
      
      // Обработчик клика по toggle
      toggle.addEventListener('click', (e) => {
        e.preventDefault();
        const isOpen = wrapper.classList.contains('open');
        
        // Закрываем все dropdown'ы
        document.querySelectorAll('.custom-select').forEach(select => {
          select.classList.remove('open');
          select.querySelector('.custom-select-menu').hidden = true;
        });
        
        // Открываем текущий, если он был закрыт
        if (!isOpen) {
          wrapper.classList.add('open');
          menu.hidden = false;
        }
      });
      
      // Обработчики для опций
      options.forEach(option => {
        option.addEventListener('click', () => {
          const value = option.getAttribute('data-value');
          const text = option.textContent;
          const textElement = toggle.querySelector('.custom-select-text');
          
          textElement.textContent = text;
          
          // Убираем выделение со всех опций
          options.forEach(opt => opt.classList.remove('selected'));
          
          // Выделяем выбранную опцию
          option.classList.add('selected');
          
          // Закрываем dropdown
          wrapper.classList.remove('open');
          menu.hidden = true;
          
          // Автоматически применяем фильтры
          if (typeof applyFilters === 'function') {
            setTimeout(applyFilters, 100);
          }
        });
      });
    });
    
    // Закрытие dropdown'ов при клике вне их
    document.addEventListener('click', (e) => {
      if (!e.target.closest('.custom-select')) {
        document.querySelectorAll('.custom-select').forEach(select => {
          select.classList.remove('open');
          select.querySelector('.custom-select-menu').hidden = true;
        });
      }
    });
  };

  // Улучшенная работа с таблицами на мобильных
  const enhanceMobileTables = () => {
    if (!isMobile()) return;

    const tables = document.querySelectorAll('.projects-table');
    
    tables.forEach(table => {
      const rows = table.querySelectorAll('.table-row');
      
      rows.forEach(row => {
        // Добавляем обработчик клика для строк таблицы
        row.addEventListener('click', () => {
          // Убираем выделение со всех строк
          rows.forEach(r => r.classList.remove('selected'));
          
          // Добавляем выделение к текущей строке
          row.classList.add('selected');
          
          // Показываем дополнительную информацию (если есть)
          showRowDetails(row);
        });
        
        // Добавляем обработчик касания для haptic feedback
        row.addEventListener('touchstart', () => {
          if (navigator.vibrate) {
            navigator.vibrate(5);
          }
        });
      });
    });
  };

  // Показать детали строки таблицы
  const showRowDetails = (row) => {
    const projectName = row.querySelector('.table-cell:first-child').textContent;
    const status = row.querySelector('.status-badge')?.textContent || '';
    const progress = row.querySelector('.progress-text')?.textContent || '';
    const budget = row.querySelector('.table-cell:nth-child(4)')?.textContent || '';
    const deadline = row.querySelector('.table-cell:last-child')?.textContent || '';
    
    // Создаем модальное окно с деталями
    const modal = document.createElement('div');
    modal.className = 'modal';
    modal.innerHTML = `
      <div class="modal-backdrop" data-close="modal"></div>
      <div class="modal-dialog">
        <div class="modal-head">
          <h3>${projectName}</h3>
        </div>
        <div class="modal-body">
          <div class="detail-item">
            <strong>Статус:</strong> ${status}
          </div>
          <div class="detail-item">
            <strong>Прогресс:</strong> ${progress}
          </div>
          <div class="detail-item">
            <strong>Бюджет:</strong> ${budget}
          </div>
          <div class="detail-item">
            <strong>Дедлайн:</strong> ${deadline}
          </div>
        </div>
        <div class="modal-actions actions">
          <button class="btn btn-outline" data-close="modal">Закрыть</button>
        </div>
      </div>
    `;
    
    // Добавляем модальное окно на страницу
    document.body.appendChild(modal);
    
    // Показываем модальное окно
    requestAnimationFrame(() => {
      modal.hidden = false;
    });
    
    // Обработчик закрытия
    modal.addEventListener('click', (e) => {
      if (e.target?.dataset?.close === 'modal') {
        modal.remove();
      }
    });
    
    // Закрытие по Escape
    const handleEscape = (e) => {
      if (e.key === 'Escape') {
        modal.remove();
        document.removeEventListener('keydown', handleEscape);
      }
    };
    document.addEventListener('keydown', handleEscape);
  };

  // Улучшенная работа с dropzone на мобильных
  const enhanceMobileDropzone = () => {
    if (!isMobile()) return;

    const dropzones = document.querySelectorAll('.dropzone');
    
    dropzones.forEach(dropzone => {
      const input = dropzone.querySelector('input[type="file"]');
      const pickBtn = dropzone.querySelector('.btn');
      
      if (!input || !pickBtn) return;
      
      // Улучшенная обработка касаний
      dropzone.addEventListener('touchstart', () => {
        dropzone.classList.add('hover');
        if (navigator.vibrate) {
          navigator.vibrate(10);
        }
      });
      
      dropzone.addEventListener('touchend', () => {
        dropzone.classList.remove('hover');
      });
      
      // Обработчик клика по dropzone
      dropzone.addEventListener('click', (e) => {
        if (e.target === dropzone || e.target === pickBtn) {
          e.preventDefault();
          input.click();
        }
      });
      
      // Обработчик изменения файлов
      input.addEventListener('change', () => {
        if (input.files && input.files.length > 0) {
          // Показываем haptic feedback
          if (navigator.vibrate) {
            navigator.vibrate(20);
          }
          
          // Обновляем UI
          updateDropzoneUI(dropzone, input.files);
        }
      });
    });
  };

  // Обновление UI dropzone
  const updateDropzoneUI = (dropzone, files) => {
    const info = dropzone.parentElement.querySelector('.upload-info');
    if (!info) return;
    
    const fileList = Array.from(files).map(f => 
      `${f.name} — ${(f.size / 1024).toFixed(1)} КБ`
    ).join('<br>');
    
    info.innerHTML = fileList || 'Файлы не выбраны';
    info.hidden = false;
    
    // Активируем кнопки
    const buttons = dropzone.parentElement.querySelectorAll('.btn:not([disabled])');
    buttons.forEach(btn => {
      if (btn.textContent.includes('Начать') || btn.textContent.includes('Обработать')) {
        btn.disabled = false;
      }
    });
  };

  // Улучшенная работа с модальными окнами на мобильных
  const enhanceMobileModals = () => {
    if (!isMobile()) return;

    // Добавляем поддержку свайпов для закрытия модальных окон
    const modals = document.querySelectorAll('.modal');
    
    modals.forEach(modal => {
      let startY = 0;
      let currentY = 0;
      
      const handleTouchStart = (e) => {
        startY = e.touches[0].clientY;
      };
      
      const handleTouchMove = (e) => {
        currentY = e.touches[0].clientY;
        const diffY = currentY - startY;
        
        if (diffY > 0) {
          // Свайп вниз - закрыть модальное окно
          modal.style.transform = `translateY(${diffY}px)`;
        }
      };
      
      const handleTouchEnd = () => {
        const diffY = currentY - startY;
        
        if (diffY > 100) {
          // Если свайп достаточно большой, закрываем модальное окно
          modal.remove();
        } else {
          // Возвращаем модальное окно на место
          modal.style.transform = '';
        }
      };
      
      modal.addEventListener('touchstart', handleTouchStart, { passive: true });
      modal.addEventListener('touchmove', handleTouchMove, { passive: true });
      modal.addEventListener('touchend', handleTouchEnd, { passive: true });
    });
  };

  // Улучшенная работа с формами на мобильных
  const enhanceMobileForms = () => {
    if (!isMobile()) return;

    const inputs = document.querySelectorAll('input, textarea, select');
    
    inputs.forEach(input => {
      // Добавляем haptic feedback при фокусе
      input.addEventListener('focus', () => {
        if (navigator.vibrate) {
          navigator.vibrate(5);
        }
      });
      
      // Улучшенная валидация в реальном времени
      input.addEventListener('input', () => {
        validateInput(input);
      });
      
      // Обработка отправки формы
      if (input.type === 'submit' || input.tagName === 'BUTTON') {
        input.addEventListener('click', (e) => {
          if (navigator.vibrate) {
            navigator.vibrate(10);
          }
        });
      }
    });
  };

  // Валидация поля ввода
  const validateInput = (input) => {
    const value = input.value.trim();
    let isValid = true;
    let errorMessage = '';
    
    // Проверяем обязательные поля
    if (input.required && !value) {
      isValid = false;
      errorMessage = 'Это поле обязательно для заполнения';
    }
    
    // Проверяем email
    if (input.type === 'email' && value && !isValidEmail(value)) {
      isValid = false;
      errorMessage = 'Введите корректный email адрес';
    }
    
    // Показываем/скрываем ошибку
    showInputError(input, isValid, errorMessage);
    
    return isValid;
  };

  // Проверка корректности email
  const isValidEmail = (email) => {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return emailRegex.test(email);
  };

  // Показать/скрыть ошибку поля
  const showInputError = (input, isValid, message) => {
    let errorElement = input.parentElement.querySelector('.input-error');
    
    if (!isValid && !errorElement) {
      errorElement = document.createElement('div');
      errorElement.className = 'input-error';
      errorElement.style.cssText = `
        color: #ef4444;
        font-size: 12px;
        margin-top: 4px;
        padding: 4px 8px;
        background: rgba(239, 68, 68, 0.1);
        border-radius: 6px;
        border: 1px solid rgba(239, 68, 68, 0.3);
      `;
      input.parentElement.appendChild(errorElement);
    }
    
    if (errorElement) {
      if (isValid) {
        errorElement.remove();
        input.style.borderColor = '';
      } else {
        errorElement.textContent = message;
        input.style.borderColor = '#ef4444';
      }
    }
  };

  // Улучшенная работа с уведомлениями на мобильных
  const enhanceMobileNotifications = () => {
    if (!isMobile()) return;

    // Переопределяем функцию toast для мобильных
    if (typeof window.toast === 'function') {
      const originalToast = window.toast;
      
      window.toast = function(msg, ok = true) {
        // Показываем haptic feedback
        if (navigator.vibrate) {
          navigator.vibrate(ok ? 10 : 20);
        }
        
        // Вызываем оригинальную функцию
        return originalToast.call(this, msg, ok);
      };
    }
  };

  // Адаптация для разных размеров экрана
  const adaptToScreenSize = () => {
    const handleResize = () => {
      // Обновляем CSS переменные в зависимости от размера экрана
      const root = document.documentElement;
      
      if (isPhone()) {
        root.style.setProperty('--mobile-scale', '0.9');
        root.style.setProperty('--touch-target-size', '44px');
      } else if (isTablet()) {
        root.style.setProperty('--mobile-scale', '0.95');
        root.style.setProperty('--touch-target-size', '40px');
      } else {
        root.style.setProperty('--mobile-scale', '1');
        root.style.setProperty('--touch-target-size', '36px');
      }
    };
    
    // Вызываем сразу и при изменении размера окна
    handleResize();
    window.addEventListener('resize', handleResize);
  };

  // Инициализация всех улучшений
  const initMobileEnhancements = () => {
    // Ждем загрузки DOM
    if (document.readyState === 'loading') {
      document.addEventListener('DOMContentLoaded', initMobileEnhancements);
      return;
    }
    
    // Применяем улучшения только на мобильных устройствах
    if (isMobile()) {
      enhanceTouchExperience();
      enhanceMobileNavigation();
      enhanceMobileFilters();
      enhanceMobileTables();
      enhanceMobileDropzone();
      enhanceMobileModals();
      enhanceMobileForms();
      enhanceMobileNotifications();
      adaptToScreenSize();
      
      console.log('Мобильные улучшения применены');
    }
  };

  // Запускаем инициализацию
  initMobileEnhancements();

  // Экспортируем функции для использования в других скриптах
  window.MobileEnhancements = {
    isMobile,
    isTablet,
    isPhone,
    enhanceTouchExperience,
    enhanceMobileNavigation,
    enhanceMobileFilters,
    enhanceMobileTables,
    enhanceMobileDropzone,
    enhanceMobileModals,
    enhanceMobileForms,
    enhanceMobileNotifications,
    adaptToScreenSize
  };

})();
