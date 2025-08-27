# gost_report.py
# pip install reportlab
from __future__ import annotations

from datetime import datetime
from typing import Any, Dict, List, Optional, Union
import json
import os
from reportlab.platypus import (
    SimpleDocTemplate, Paragraph, Spacer, PageBreak,
    ListFlowable, ListItem, Table, TableStyle   # + Table, TableStyle
)
from reportlab.lib import colors               # + colors

from reportlab.lib.pagesizes import A4
from reportlab.lib.units import mm
from reportlab.lib.styles import ParagraphStyle, getSampleStyleSheet
from reportlab.lib.enums import TA_JUSTIFY, TA_CENTER, TA_LEFT
from reportlab.platypus import (
    SimpleDocTemplate, Paragraph, Spacer, PageBreak,
    ListFlowable, ListItem
)
from reportlab.platypus.tableofcontents import TableOfContents
from reportlab.pdfbase import pdfmetrics
from reportlab.pdfbase.ttfonts import TTFont
from reportlab.pdfgen import canvas

from reportlab.pdfbase.ttfonts import TTFont
from reportlab.pdfbase import pdfmetrics

def _register_times_new_roman():
    # подключаем шрифт с кириллицей
    pdfmetrics.registerFont(TTFont("DejaVuSerif", "DejaVuSerif.ttf"))
    pdfmetrics.registerFont(TTFont("DejaVuSerif-Bold", "DejaVuSerif-Bold.ttf"))
    return {
        "regular": "DejaVuSerif",
        "bold": "DejaVuSerif-Bold",
        "italic": "DejaVuSerif",  # можно заменить на Italic, если есть
    }

# ---------- шрифты ----------


# ---------- номер страницы внизу ----------
def _page_number_footer(c: canvas.Canvas, doc):
    c.setFont(doc._fonts["regular"], 10)
    text = str(c.getPageNumber())
    w = c.stringWidth(text, doc._fonts["regular"], 10)
    c.drawString((doc.pagesize[0] - w) / 2.0, 12 * mm, text)


# ---------- свой документ, чтобы хранить выбранные шрифты ----------
class GostDoc(SimpleDocTemplate):
    def __init__(self, filename, **kwargs):
        super().__init__(filename, **kwargs)
        self._fonts = _register_times_new_roman()


# ---------- стили ----------
def _build_styles(fonts: Dict[str, str]):
    from reportlab.lib.styles import ParagraphStyle, getSampleStyleSheet

    def set_or_update(styles, style: ParagraphStyle):
        # если стиль уже есть – обновим поля, иначе добавим
        if style.name in styles.byName:
            old = styles.byName[style.name]
            for k, v in style.__dict__.items():
                # только публичные атрибуты ParagraphStyle
                if not k.startswith('_') and hasattr(old, k):
                    setattr(old, k, v)
        else:
            styles.add(style)

    styles = getSampleStyleSheet()

    styles.add(ParagraphStyle(
    name="GOST-Body-Bold",
    fontName=fonts["bold"],
    fontSize=14,
    leading=21,
    alignment=TA_JUSTIFY,
    spaceAfter=6,
))
    # базовые
    set_or_update(styles, ParagraphStyle(
        name="GOST-Body",
        fontName=fonts["regular"], fontSize=14, leading=21,
        alignment=TA_JUSTIFY, spaceAfter=6,
    ))
    set_or_update(styles, ParagraphStyle(
        name="GOST-H1",
        fontName=fonts["bold"], fontSize=16, leading=22,
        alignment=TA_CENTER, spaceBefore=12, spaceAfter=12, keepWithNext=True,
    ))
    # подзаголовок с отступом
    set_or_update(styles, ParagraphStyle(
        name="GOST-H2",
        fontName=fonts["bold"], fontSize=14, leading=21,
        alignment=TA_LEFT, leftIndent=8 * mm,  # <- отступ
        spaceBefore=12, spaceAfter=8, keepWithNext=True,
    ))
    set_or_update(styles, ParagraphStyle(
        name="GOST-TOC-Title",
        fontName=fonts["bold"], fontSize=16, leading=22,
        alignment=TA_CENTER, spaceAfter=12,
    ))
    set_or_update(styles, ParagraphStyle(
        name="GOST-TitleBig",
        fontName=fonts["bold"], fontSize=18, leading=24,
        alignment=TA_CENTER, spaceAfter=12,
    ))
    set_or_update(styles, ParagraphStyle(
        name="GOST-TitleMid",
        fontName=fonts["regular"], fontSize=14, leading=20,
        alignment=TA_CENTER, spaceAfter=6,
    ))
    set_or_update(styles, ParagraphStyle(
        name="GOST-TitleSmall",
        fontName=fonts["regular"], fontSize=12, leading=16,
        alignment=TA_CENTER, spaceAfter=0,
    ))
    set_or_update(styles, ParagraphStyle(
        name="GOST-List",
        fontName=fonts["regular"], fontSize=14, leading=21,
        leftIndent=14, spaceAfter=3,
    ))
    # стили для оглавления
    set_or_update(styles, ParagraphStyle(
        name="TOCHeading1",
        fontName=fonts["regular"], fontSize=14, leading=21,
        leftIndent=0, firstLineIndent=0, spaceAfter=4,
    ))
    set_or_update(styles, ParagraphStyle(
        name="TOCHeading2",
        fontName=fonts["regular"], fontSize=12, leading=18,
        leftIndent=14, firstLineIndent=0, spaceAfter=2,
    ))
    return styles



# ---------- утилиты верстки ----------
def _safe_para(text: str, style) -> Paragraph:
    """Экранируем &, <, > чтобы reportlab не сломался на 'html'. """
    t = (text or "").replace("&", "&amp;").replace("<", "&lt;").replace(">", "&gt;")
    return Paragraph(t, style)


def _after_flowable_add_toc(doc: GostDoc, level: int, text: str):
    """Добавляем запись в TOC после вывода заголовка."""
    def _callback(canv, _doc):
        _doc.notify('TOCEntry', (level, text, _doc.page))
    doc.afterFlowable(_callback)


def _title_page(story: List, styles, meta: Dict[str, Any]):
    customer = meta.get("customer", "ПАО «Газпром»")
    executor = meta.get("executor", "")
    doc_code = meta.get("doc_code", "")
    city = meta.get("city", "Москва")
    year = meta.get("year", datetime.now().year)
    title = meta.get("title", "Отчёт по результатам анализа")

    story.append(Spacer(1, 40 * mm))
    story.append(_safe_para(customer, styles["GOST-TitleMid"]))
    story.append(Spacer(1, 10 * mm))
    story.append(_safe_para(title, styles["GOST-TitleBig"]))
    if doc_code:
        story.append(_safe_para(f"Шифр документа: {doc_code}", styles["GOST-TitleMid"]))
    if executor:
        story.append(Spacer(1, 20 * mm))
        story.append(_safe_para(f"Исполнитель: {executor}", styles["GOST-TitleMid"]))

    story.append(Spacer(1, 60 * mm))
    story.append(_safe_para(str(city), styles["GOST-TitleSmall"]))
    story.append(_safe_para(str(year), styles["GOST-TitleSmall"]))
    story.append(PageBreak())


def _make_toc(story: List, styles):
    story.append(_safe_para("СОДЕРЖАНИЕ", styles["GOST-TOC-Title"]))
    toc = TableOfContents()
    toc.levelStyles = [styles["TOCHeading1"], styles["TOCHeading2"]]
    story.append(toc)
    story.append(PageBreak())


def _section(story: List, styles, doc: GostDoc, title: str, level: int):
    style = styles["GOST-H1"] if level == 0 else styles["GOST-H2"]
    story.append(_safe_para(title, style))
    _after_flowable_add_toc(doc, level, title)


def _remarks_list(remarks: List[str], styles) -> ListFlowable:
    items = [
        ListItem(_safe_para((txt or "").strip(), styles["GOST-List"]), value=i)
        for i, txt in enumerate(remarks or [], start=1)
    ]
    return ListFlowable(items, bulletType='1', start='1', leftIndent=18)

def _remarks_table(remarks: List[str], styles, fonts) -> Table:
    """
    Таблица: № | Замечание
    Длинные тексты автоматически переносятся (через Paragraph).
    """
    # Заголовок
    data = [
        [_safe_para("№", styles["GOST-Body"]),
         _safe_para("Замечание", styles["GOST-Body"])]
    ]
    # Строки
    for i, txt in enumerate(remarks or [], start=1):
        data.append([
            _safe_para(str(i), styles["GOST-Body"]),
            _safe_para((txt or "").strip(), styles["GOST-Body"])
        ])

    tbl = Table(
        data,
        colWidths=[12 * mm, None],   # фиксируем узкую колонку под номер
        hAlign="LEFT"
    )
    tbl.setStyle(TableStyle([
        # сетка
        ("GRID", (0, 0), (-1, -1), 0.5, colors.black),
        # фон заголовка
        ("BACKGROUND", (0, 0), (-1, 0), colors.whitesmoke),
        # жирный шрифт для заголовка
        ("FONTNAME", (0, 0), (-1, 0), fonts["bold"]),
        # выравнивания
        ("ALIGN", (0, 0), (0, -1), "CENTER"),
        ("VALIGN", (0, 0), (-1, -1), "TOP"),
        # внутренние отступы
        ("LEFTPADDING", (0, 0), (-1, -1), 4),
        ("RIGHTPADDING", (0, 0), (-1, -1), 4),
        ("TOPPADDING", (0, 0), (-1, -1), 3),
        ("BOTTOMPADDING", (0, 0), (-1, -1), 3),
    ]))
    return tbl


# ---------- основная функция ----------
def json_to_gost_pdf(
    json_input: Union[str, List[Dict[str, Any]]],
    output_pdf_path: str,
    meta: Optional[Dict[str, Any]] = None,
    add_conclusion: bool = True,
):
    """
    Преобразует JSON (путь к файлу или уже загруженный list[dict]) в PDF-отчёт (ГОСТ-стиль).
    Ожидаемая структура элемента:
      {
        "category": "Раздел",
        "items": [
          {
            "group_name": "Подраздел",
            "synthesized_text": "Текст",
            "original_remarks": ["...", "..."]
          }
        ]
      }
    """
    # загрузка
    if isinstance(json_input, str):
        with open(json_input, "r", encoding="utf-8") as f:
            data = json.load(f)
    else:
        data = json_input

    meta = meta or {}

    # параметры листа/полей (ГОСТ-подобные)
    doc = GostDoc(
        output_pdf_path,
        pagesize=A4,
        leftMargin=30 * mm,   # слева 30 мм
        rightMargin=10 * mm,  # справа 10 мм
        topMargin=20 * mm,
        bottomMargin=20 * mm,
        title=meta.get("title", "Отчёт"),
        author=meta.get("executor", ""),
        subject=meta.get("customer", "ПАО «Газпром»"),
    )
    styles = _build_styles(doc._fonts)
    story: List = []

    # титульный
    _title_page(story, styles, meta)

    # оглавление
    _make_toc(story, styles)

    # введение
    _section(story, styles, doc, "ВВЕДЕНИЕ", level=0)
    intro_text = meta.get(
        "intro",
        "Настоящий отчёт подготовлен на основании предоставленных данных в формате JSON. "
        "Категории данных сформированы как главы, группы — как подразделы. "
        "Для каждого подраздела приведены синтезированное описание и исходные замечания."
    )
    story.append(_safe_para(intro_text, styles["GOST-Body"]))
    story.append(Spacer(1, 4 * mm))

    # основные разделы
    for block in data:
        category = block.get("category") or "Раздел"
        items = block.get("items") or []

        _section(story, styles, doc, category, level=0)
        story.append(Spacer(1, 4 * mm))

        for it in items:
            group = it.get("group_name") or "Подраздел"
            synth = it.get("synthesized_text") or ""
            remarks = it.get("original_remarks") or []

            _section(story, styles, doc, group, level=1)

            if synth:
                story.append(
                    Paragraph("Краткая сводка: ", styles["GOST-Body-Bold"])
                )
                story.append(
                    Paragraph(synth, styles["GOST-Body"])
                )
            if remarks:
                story.append(_remarks_table(remarks, styles, doc._fonts))
                story.append(Spacer(1, 4 * mm))
                
        story.append(Spacer(1, 6 * mm))

    # заключение
    if add_conclusion:
        story.append(PageBreak())
        _section(story, styles, doc, "ЗАКЛЮЧЕНИЕ", level=0)
        conclusion = meta.get(
            "conclusion",
            "Предложенные мероприятия направлены на снижение неопределённостей и повышение качества "
            "прогнозов. Рекомендовано согласовать план доизучения и актуализировать модели по итогам "
            "получения новых данных."
        )
        story.append(_safe_para(conclusion, styles["GOST-Body"]))

    # сборка
    doc.multiBuild(story, onFirstPage=_page_number_footer, onLaterPages=_page_number_footer)

# --------- пример самостоятельного запуска (необязательно) ---------
if __name__ == "__main__":
    # пример использования
    input_json_path = "synthesis_report_clustered.json"   # поменяйте на свой путь
    out_pdf = "gazprom_report.pdf"
    meta = {
        "customer": "ПАО «Газпром»",
        "executor": "Отдел аналитики",
        "doc_code": "GAZ-REP-2025-001",
        "city": "Москва",
        "year": 2025,
        "title": "Сводный отчёт по замечаниям и программе доизучения",
    }
    #print("asd")
    json_to_gost_pdf(input_json_path, out_pdf, meta)
    print("Готово:", out_pdf)
