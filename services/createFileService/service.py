from __future__ import annotations
from typing import Any, Dict, List, Optional
from datetime import datetime
from io import BytesIO
import os

from fastapi import FastAPI, Body, Response
from fastapi.responses import JSONResponse
from fastapi.middleware.cors import CORSMiddleware
import uvicorn
import os

from reportlab.lib.pagesizes import A4
from reportlab.lib.units import mm
from reportlab.lib.styles import ParagraphStyle, getSampleStyleSheet
from reportlab.lib.enums import TA_JUSTIFY, TA_CENTER, TA_LEFT
from reportlab.platypus import SimpleDocTemplate, Paragraph, Spacer, PageBreak, Table, TableStyle
from reportlab.platypus.tableofcontents import TableOfContents
from reportlab.lib import colors
from reportlab.pdfbase import pdfmetrics
from reportlab.pdfbase.ttfonts import TTFont
from reportlab.pdfgen import canvas


app = FastAPI()

async def main():
    # Настройка CORS
    app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
    )

# Эндпоинт для проверки работы сервера
@app.get("/health")
async def health_check():
    return {"status": "ok"}

def build_gost_pdf_bytes(
    json_obj: List[Dict[str, Any]],
    meta: Optional[Dict[str, Any]] = None,
    add_conclusion: bool = True,
) -> bytes:
    """
    Генерирует PDF (байты) из JSON-объекта формата:
    [
      {
        "category": "Раздел",
        "items": [
          {
            "group_name": "Подраздел",
            "synthesized_text": "Краткая сводка (одна-две фразы)",
            "original_remarks": ["замечание 1", "замечание 2", ...]
          }, ...
        ]
      }, ...
    ]

    meta: {
      "customer": "ПАО «Газпром»",
      "title": "Сводный отчёт ...",
      "executor": "Отдел аналитики",
      "city": "Москва",
      "year": 2025,
      "doc_code": "GAZ-REP-...",
      "intro": "...",        # опционально
      "conclusion": "..."    # опционально
    }
    """
    meta = meta or {}

    # ---------- утилиты внутри функции ----------
    def _safe(t: str) -> str:
        return (t or "").replace("&", "&amp;").replace("<", "&lt;").replace(">", "&gt;")

    def _register_fonts() -> Dict[str, str]:
        # 1) ENV пути к ttf (надёжно для контейнера)
        reg_env = os.getenv("PDF_FONT_REGULAR")
        bold_env = os.getenv("PDF_FONT_BOLD")
        if reg_env and os.path.exists(reg_env):
            pdfmetrics.registerFont(TTFont("App-Regular", reg_env))
            if bold_env and os.path.exists(bold_env):
                pdfmetrics.registerFont(TTFont("App-Bold", bold_env))
                return {"regular": "App-Regular", "bold": "App-Bold"}
            return {"regular": "App-Regular", "bold": "App-Regular"}

        # 2) Стандартные Linux шрифты (DejaVu, Liberation, Noto)
        linux_font_paths = [
            # DejaVu (часто установлены по умолчанию)
            "/usr/share/fonts/truetype/dejavu/DejaVuSerif.ttf",
            "/usr/share/fonts/dejavu/DejaVuSerif.ttf",
            # Liberation (альтернатива Arial)
            "/usr/share/fonts/liberation/LiberationSerif-Regular.ttf",
            "/usr/share/fonts/truetype/liberation/LiberationSerif-Regular.ttf",
            # Noto (современные шрифты от Google)
            "/usr/share/fonts/truetype/noto/NotoSerif-Regular.ttf",
            # Ubuntu  
            "/usr/share/fonts/truetype/ubuntu/Ubuntu-R.ttf",
            # Общий путь для многих дистрибутивов
            "/usr/share/fonts/truetype/freefont/FreeSerif.ttf"
        ]
        
        bold_paths = [
            "/usr/share/fonts/truetype/dejavu/DejaVuSerif-Bold.ttf",
            "/usr/share/fonts/dejavu/DejaVuSerif-Bold.ttf",
            "/usr/share/fonts/liberation/LiberationSerif-Bold.ttf",
            "/usr/share/fonts/truetype/liberation/LiberationSerif-Bold.ttf",
            "/usr/share/fonts/truetype/noto/NotoSerif-Bold.ttf",
            "/usr/share/fonts/truetype/ubuntu/Ubuntu-B.ttf",
            "/usr/share/fonts/truetype/freefont/FreeSerifBold.ttf"
        ]

        # Проверяем доступные шрифты
        for reg_font in linux_font_paths:
            if os.path.exists(reg_font):
                font_name = os.path.splitext(os.path.basename(reg_font))[0]
                pdfmetrics.registerFont(TTFont(font_name, reg_font))
                
                # Ищем соответствующий bold-шрифт
                bold_font = None
                for b_font in bold_paths:
                    if os.path.exists(b_font):
                        bold_name = os.path.splitext(os.path.basename(b_font))[0]
                        pdfmetrics.registerFont(TTFont(bold_name, b_font))
                        return {"regular": font_name, "bold": bold_name}
                
                return {"regular": font_name, "bold": font_name}

        # 3) Локальные шрифты рядом с кодом
        if os.path.exists("DejaVuSerif.ttf"):
            pdfmetrics.registerFont(TTFont("DejaVuSerif", "DejaVuSerif.ttf"))
            if os.path.exists("DejaVuSerif-Bold.ttf"):
                pdfmetrics.registerFont(TTFont("DejaVuSerif-Bold", "DejaVuSerif-Bold.ttf"))
                return {"regular": "DejaVuSerif", "bold": "DejaVuSerif-Bold"}
            return {"regular": "DejaVuSerif", "bold": "DejaVuSerif"}

        # 4) Fallback (может не поддерживать кириллицу)
        return {"regular": "Times-Roman", "bold": "Times-Bold"}

    def _footer(c: canvas.Canvas, doc):
        c.setFont(fonts["regular"], 10)
        s = str(c.getPageNumber())
        w = c.stringWidth(s, fonts["regular"], 10)
        c.drawString((doc.pagesize[0]-w)/2, 12*mm, s)

    def _styles(fonts: Dict[str, str]):
        ss = getSampleStyleSheet()

        def set_or_update(ps: ParagraphStyle):
            if ps.name in ss.byName:
                old = ss.byName[ps.name]
                for k, v in ps.__dict__.items():
                    if not k.startswith("_") and hasattr(old, k):
                        setattr(old, k, v)
            else:
                ss.add(ps)

        set_or_update(ParagraphStyle(name="GOST-Body", fontName=fonts["regular"], fontSize=14,
                                     leading=21, alignment=TA_JUSTIFY, spaceAfter=6))
        set_or_update(ParagraphStyle(name="GOST-Body-Bold", fontName=fonts["bold"], fontSize=14,
                                     leading=21, alignment=TA_JUSTIFY, spaceAfter=6))
        set_or_update(ParagraphStyle(name="GOST-H1", fontName=fonts["bold"], fontSize=16,
                                     leading=22, alignment=TA_CENTER, spaceBefore=12, spaceAfter=12, keepWithNext=True))
        set_or_update(ParagraphStyle(name="GOST-H2", fontName=fonts["bold"], fontSize=14,
                                     leading=21, alignment=TA_LEFT, leftIndent=8*mm, spaceBefore=12, spaceAfter=8, keepWithNext=True))
        set_or_update(ParagraphStyle(name="GOST-TOC", fontName=fonts["bold"], fontSize=16,
                                     leading=22, alignment=TA_CENTER, spaceAfter=12))
        set_or_update(ParagraphStyle(name="TOC1", fontName=fonts["regular"], fontSize=14, leading=21,
                                     leftIndent=0, firstLineIndent=0, spaceAfter=4))
        set_or_update(ParagraphStyle(name="TOC2", fontName=fonts["regular"], fontSize=12, leading=18,
                                     leftIndent=14, firstLineIndent=0, spaceAfter=2))
        set_or_update(ParagraphStyle(name="TitleBig", fontName=fonts["bold"], fontSize=18,
                                     leading=24, alignment=TA_CENTER, spaceAfter=12))
        set_or_update(ParagraphStyle(name="TitleMid", fontName=fonts["regular"], fontSize=14,
                                     leading=20, alignment=TA_CENTER, spaceAfter=6))
        set_or_update(ParagraphStyle(name="TitleSmall", fontName=fonts["regular"], fontSize=12,
                                     leading=16, alignment=TA_CENTER, spaceAfter=0))
        return ss

    def _add_toc_entry(doc, level: int, text: str):
        def cb(canv, _doc): _doc.notify('TOCEntry', (level, text, _doc.page))
        doc.afterFlowable(cb)

    def _table_summary(summary_text: str):
        data = [
            [Paragraph(_safe("Краткая сводка:"), styles["GOST-Body-Bold"]),
             Paragraph(_safe(summary_text), styles["GOST-Body"])]
        ]
        t = Table(data, colWidths=[45*mm, None], hAlign="LEFT")
        t.setStyle(TableStyle([
            ("VALIGN", (0,0), (-1,-1), "TOP"),
            ("LEFTPADDING", (0,0), (-1,-1), 0),
            ("RIGHTPADDING", (0,0), (-1,-1), 0),
            ("TOPPADDING", (0,0), (-1,-1), 0),
            ("BOTTOMPADDING", (0,0), (-1,-1), 3),
        ]))
        return t

    def _table_remarks(remarks: List[str]):
        data = [
            [Paragraph(_safe("№"), styles["GOST-Body-Bold"]),
             Paragraph(_safe("Замечание"), styles["GOST-Body-Bold"])]
        ]
        for i, txt in enumerate(remarks or [], start=1):
            data.append([
                Paragraph(str(i), styles["GOST-Body"]),
                Paragraph(_safe((txt or "").strip()), styles["GOST-Body"])
            ])
        t = Table(data, colWidths=[12*mm, None], hAlign="LEFT")
        t.setStyle(TableStyle([
            ("GRID", (0,0), (-1,-1), 0.5, colors.black),
            ("BACKGROUND", (0,0), (-1,0), colors.whitesmoke),
            ("FONTNAME", (0,0), (-1,0), fonts["bold"]),
            ("ALIGN", (0,0), (0,-1), "CENTER"),
            ("VALIGN", (0,0), (-1,-1), "TOP"),
            ("LEFTPADDING", (0,0), (-1,-1), 4),
            ("RIGHTPADDING", (0,0), (-1,-1), 4),
            ("TOPPADDING", (0,0), (-1,-1), 3),
            ("BOTTOMPADDING", (0,0), (-1,-1), 3),
        ]))
        return t

    # ---------- сборка PDF ----------
    fonts = _register_fonts()
    buf = BytesIO()

    class _Doc(SimpleDocTemplate):
        def __init__(self, *a, **kw):
            super().__init__(*a, **kw)

    doc = _Doc(
        buf,
        pagesize=A4,
        leftMargin=30*mm, rightMargin=10*mm, topMargin=20*mm, bottomMargin=20*mm,
        title=meta.get("title", "Отчёт"),
        author=meta.get("executor", ""),
        subject=meta.get("customer", "ПАО «Газпром»"),
    )
    styles = _styles(fonts)
    story: List[Any] = []

    # титульник
    story.append(Spacer(1, 40*mm))
    story.append(Paragraph(_safe(meta.get("customer","ПАО «Газпром»")), styles["TitleMid"]))
    story.append(Spacer(1, 10*mm))
    story.append(Paragraph(_safe(meta.get("title","Сводный отчёт")), styles["TitleBig"]))
    if meta.get("doc_code"):
        story.append(Paragraph(_safe(f"Шифр документа: {meta['doc_code']}"), styles["TitleMid"]))
    if meta.get("executor"):
        story.append(Spacer(1, 20*mm))
        story.append(Paragraph(_safe(f"Исполнитель: {meta['executor']}"), styles["TitleMid"]))
    story.append(Spacer(1, 60*mm))
    story.append(Paragraph(_safe(meta.get("city","Москва")), styles["TitleSmall"]))
    story.append(Paragraph(_safe(str(meta.get("year", datetime.now().year))), styles["TitleSmall"]))
    story.append(PageBreak())

    # оглавление
    story.append(Paragraph("СОДЕРЖАНИЕ", styles["GOST-TOC"]))
    toc = TableOfContents(); toc.levelStyles = [styles["TOC1"], styles["TOC2"]]
    story.append(toc); story.append(PageBreak())

    # введение
    story.append(Paragraph("ВВЕДЕНИЕ", styles["GOST-H1"])); _add_toc_entry(doc, 0, "ВВЕДЕНИЕ")
    intro = meta.get("intro",
        "Отчёт сформирован на основании данных JSON: категории представлены как главы, "
        "группы — как подразделы. Для каждого подраздела приводится «Краткая сводка» и "
        "таблица исходных замечаний.")
    story.append(Paragraph(_safe(intro), styles["GOST-Body"]))
    story.append(Spacer(1, 4*mm))

    # for block in (json_obj or []):
    #     cat = block.get("category") or "Раздел"
    #     items = block.get("items") or []

    #     story.append(Paragraph(_safe(cat), styles["GOST-H1"])); _add_toc_entry(doc, 0, cat)
    #     story.append(Spacer(1, 4*mm))

    #     for it in items:
    #         grp = it.get("group_name") or "Подраздел"
    #         synth = it.get("synthesized_text") or ""
    #         remarks = it.get("original_remarks") or []

    #         story.append(Paragraph(_safe(grp), styles["GOST-H2"])); _add_toc_entry(doc, 1, grp)

    #         if synth:
    #             story.append(_table_summary(synth))
    #             story.append(Spacer(1, 2*mm))
    #         if remarks:
    #             story.append(_table_remarks(remarks))
    #             story.append(Spacer(1, 4*mm))

    #     story.append(Spacer(1, 6*mm))

    for section_name, section_data in (json_obj.items() if isinstance(json_obj, dict) else []):
        # Добавляем заголовок раздела  
        section_title = section_name.capitalize() if section_name else "Раздел"
        story.append(Paragraph(_safe(section_title), styles["GOST-H1"]))
        _add_toc_entry(doc, 0, section_title)
        story.append(Spacer(1, 4*mm))

        # Обрабатываем группы в разделе
        for group in (section_data if isinstance(section_data, list) else []):
            group_name = group.get("group_name") or "Подраздел"
            synthesized = group.get("synthesized_remark") or ""
            remarks = group.get("original_duplicates") or []

            # Добавляем подзаголовок группы
            story.append(Paragraph(_safe(group_name), styles["GOST-H2"]))
            _add_toc_entry(doc, 1, group_name)

            # Добавляем синтезированное замечание
            if synthesized:
                story.append(_table_summary(synthesized))
                story.append(Spacer(1, 2*mm))

            # Добавляем оригинальные замечания
            if remarks:
                story.append(_table_remarks(remarks))
                story.append(Spacer(1, 4*mm))

        # Добавляем отступ после раздела
        story.append(Spacer(1, 6*mm))

    # заключение
    if add_conclusion:
        story.append(PageBreak())
        story.append(Paragraph("ЗАКЛЮЧЕНИЕ", styles["GOST-H1"])); _add_toc_entry(doc, 0, "ЗАКЛЮЧЕНИЕ")
        conclusion = meta.get("conclusion",
                              "Предложенные мероприятия направлены на снижение неопределённостей "
                              "и повышение точности прогнозов.")
        story.append(Paragraph(_safe(conclusion), styles["GOST-Body"]))

    # двухпроходная сборка — корректное оглавление
    doc.multiBuild(story, onFirstPage=_footer, onLaterPages=_footer)
    return buf.getvalue()
    

    print("Готово: test_report.pdf")



@app.post("/remarks_report")
async def remarksHandler(
    data: dict = Body(...)):
    print(data)
    # Пример JSON (мини-версия)
    # json_obj = [
    #     {
    #         "category": "Разработка месторождения",
    #         "items": [
    #             {
    #                 "group_name": "Сейсморазведка",
    #                 "synthesized_text": "Необходимо выполнить уточнение параметров коллектора.",
    #                 "original_remarks": [
    #                     "Недостаточный охват 3D-сейсмикой",
    #                     "Не проведён анализ аномалий"
    #                 ]
    #             },
    #             {
    #                 "group_name": "Моделирование",
    #                 "synthesized_text": "Следует обновить гидродинамическую модель по новым скважинам.",
    #                 "original_remarks": [
    #                     "Нет истории эксплуатации для части фонда"
    #                 ]
    #             }
    #         ]
    #     }
    # ]

    meta = {
        "customer": "ПАО «Газпром»",
        "title": "Сводный отчёт по замечаниям",
        "executor": "Отдел аналитики",
        "city": "Москва",
        "year": 2025,
        "doc_code": "GAZ-REP-2025-001"
    }

    pdf_bytes = build_gost_pdf_bytes(data, meta=meta)

    # Сохраняем результат в файл
    #with open("test_report.pdf", "wb") as f:
    #    f.write(pdf_bytes)

    #return JSONResponse(content=pdf_bytes, status_code=200)
    return Response(
        content=pdf_bytes,
        media_type="application/pdf",
        headers={"Content-Disposition": "filename=report.pdf"},
        status_code=202
    )

if __name__ == "__main__":
    #projects_to_process = ["Project_Alfa"]
    uvicorn.run(app, host="127.0.0.1", port=8086)