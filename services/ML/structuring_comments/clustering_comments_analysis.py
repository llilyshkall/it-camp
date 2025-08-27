import torch
from sentence_transformers import SentenceTransformer, util
import json
import asyncio
from collections import defaultdict
import aiohttp
from asyncio import Semaphore
from sklearn.cluster import AgglomerativeClustering
from sklearn.metrics.pairwise import cosine_similarity
import numpy as np
import itertools

# ---  КОНФИГУРАЦИЯ ---
LOCAL_API_URL = "http://89.108.116.240:11434/api/chat"
LOCAL_MODEL_NAME = "qwen3-8b:latest"
MAX_CONCURRENT_REQUESTS = 3

# Если в кластере больше замечаний, чем это число, будет запущена выборка лучших, чтобы не перегревать модель.
MAX_REMARKS_FOR_SYNTHESIS = 10

EMBEDDING_MODEL = "intfloat/multilingual-e5-large"
DATA_FILE = "Dirty.json"
THEMES_FILE = "themes.json"
CLUSTER_DISTANCE_THRESHOLD = 0.18
NLP_CLASSIFICATION_THRESHOLD = 0.75

# ---  ИНИЦИАЛИЗАЦИЯ ---
print("Загрузка локальной модели для векторизации...")
device = "cuda" if torch.cuda.is_available() else "cpu"
embedding_model = SentenceTransformer(EMBEDDING_MODEL, device=device)
print(f"Модель для векторизации загружена на '{device}'.")

# ---  ПРОМПТЫ ---
CLASSIFY_MAJOR_PROMPT = lambda text, categories: [{"role": "system",
                                                   "content": "Ты — эксперт-аналитик... Классифицируй замечание, выбрав ОДНУ ОСНОВНУЮ категорию. Ответь НА РУССКОМ ЯЗЫКЕ ТОЛЬКО названием категории или 'Прочее'."},
                                                  {"role": "user",
                                                   "content": f"СПИСОК КАТЕГОРИЙ:\n{categories}\n\nЗАМЕЧАНИЕ:\n\"{text}\"\n\nКАТЕГОРИЯ:"}]

CLASSIFY_SUB_PROMPT = lambda text, sub_categories: [{"role": "system",
                                                     "content": "Ты — эксперт-аналитик... Классифицируй замечание, выбрав ОДНУ ПОДКАТЕГОРИЮ НА РУССКОМ ЯЗЫКЕ. Если ни одна не подходит, ответь 'None'."},
                                                    {"role": "user",
                                                     "content": f"СПИСОК ПОДКАТЕГОРИЙ:\n{sub_categories}\n\nЗАМЕЧАНИЕ:\n\"{text}\"\n\nПОДКАТЕГОРИЯ:"}]

CREATE_NEW_SUB_PROMPT = lambda text: [{"role": "system",
                                       "content": "Ты — эксперт-аналитик... Сформулируй ОДНО краткое название новой подкатегории для этого замечания. Ответь ТОЛЬКО названием НА РУССКОМ ЯЗЫКЕ."},
                                      {"role": "user",
                                       "content": f"ЗАМЕЧАНИЕ:\n\"{text}\"\n\nНАЗВАНИЕ НОВОЙ ПОДКАТЕГОРИИ:"}]

COMPLEX_SYNTHESIS_PROMPT = lambda texts: [{"role": "system",
                                           "content": "Ты — главный эксперт-аналитик... Проанализируй список схожих замечаний. Верни ответ СТРОГО в формате JSON с двумя ключами: 'group_name' (краткое название) и 'synthesized_remark' (обобщающее замечание) НА РУССКОМ ЯЗЫКЕ."},
                                          {"role": "user", "content": f"СПИСОК ЗАМЕЧАНИЙ:\n{texts}\n\nJSON ОТВЕТ:"}]


# ---  ФУНКЦИИ ПАЙПЛАЙНА ---

def load_knowledge_base(themes_file):
    try:
        with open(themes_file, 'r', encoding='utf-8') as f:
            themes = json.load(f)
        print(" База знаний тем успешно загружена.")
        # ### ИЗМЕНЕНИЕ: Убедимся, что возвращаем списки, даже если ключей нет ###
        return themes.get('major_categories', []), themes.get('sub_categories', [])
    except Exception as e:
        print(f" Ошибка загрузки базы знаний: {e}");
        return [], []


def save_knowledge_base(themes_file, major_categories, sub_categories):
    try:
        with open(themes_file, 'w', encoding='utf-8') as f:
            json.dump({"major_categories": major_categories, "sub_categories": sub_categories}, f, ensure_ascii=False,
                      indent=2)
        print("База знаний успешно обновлена и сохранена.")
    except Exception as e:
        print(f" Ошибка сохранения базы знаний: {e}")


# Разделяет замечания на две группы: уже классифицированные и те, что в "None"
def load_and_partition_remarks(file_path):
    preclassified_remarks = defaultdict(list)
    unclassified_remarks = []
    unique_texts = set()

    try:
        with open(file_path, 'r', encoding='utf-8') as f:
            data = json.load(f)

        # Сначала загрузим названия категорий из ключа 'keys'
        category_names = data.get("keys", {})

        for cat_key, remarks in data.items():
            if cat_key in ["keys", "None"]:
                continue

            # Получаем  имя категории, если оно есть
            major_category_name = category_names.get(cat_key, cat_key)

            for remark_text in remarks:
                stripped_text = remark_text.strip()
                if stripped_text and stripped_text not in unique_texts:
                    unique_texts.add(stripped_text)
                    preclassified_remarks[major_category_name].append(
                        {"id": len(unique_texts) - 1, "text": stripped_text})

        if "None" in data:
            for remark_text in data["None"]:
                stripped_text = remark_text.strip()
                if stripped_text and stripped_text not in unique_texts:
                    unique_texts.add(stripped_text)
                    unclassified_remarks.append({"id": len(unique_texts) - 1, "text": stripped_text})

        print(
            f" Данные загружены. Найдено {len(preclassified_remarks.keys())} предварительно классифицированных категорий.")
        print(f" Найдено {len(unclassified_remarks)} неклассифицированных замечаний.")
        print(f" Всего уникальных замечаний: {len(unique_texts)}.")

        return preclassified_remarks, unclassified_remarks

    except Exception as e:
        print(f" Ошибка загрузки и разделения данных: {e}")
        return {}, []


def cluster_remarks(remarks_list, embeddings, distance_threshold=0.3):
    if len(remarks_list) <= 1:
        # Если замечание одно, возвращаем кластер из одного элемента
        return [[remarks_list[0]]]

    print(f"\n--- Этап: Семантическая кластеризация (порог: {distance_threshold}) ---")
    clustering = AgglomerativeClustering(n_clusters=None, distance_threshold=distance_threshold, metric='cosine',
                                         linkage='complete').fit(embeddings)
    clusters = defaultdict(list)
    for i, cluster_id in enumerate(clustering.labels_):
        clusters[cluster_id].append(remarks_list[i])

    final_clusters = list(clusters.values())
    print(f" Замечания сгруппированы в {len(final_clusters)} семантических кластеров.")
    return final_clusters


async def process_llm_requests(items, prompt_function, semaphore, **kwargs):
    session_id = kwargs.pop("session_id", "REQ")
    async with aiohttp.ClientSession() as session:
        async def get_one(item):
            async with semaphore:
                prompt_args = {k: v for k, v in kwargs.items()}
                prompt_input_for_log = ""

                if 'texts_list' in item:
                    texts_for_prompt = "\n".join([f"- {text}" for text in item['texts_list']])
                    prompt_args['texts'] = texts_for_prompt
                    prompt_input_for_log = f"texts list of size {len(item['texts_list'])}"
                elif 'text' in item:
                    prompt_args['text'] = item['text']
                    prompt_input_for_log = item['text'][:70].strip()
                else:
                    return item, None

                messages = prompt_function(**prompt_args)
                payload = {"model": LOCAL_MODEL_NAME, "messages": messages, "stream": False}

                try:
                    async with session.post(LOCAL_API_URL, json=payload, timeout=300) as response:
                        if response.status != 200:
                            error_text = await response.text()
                            print(
                                f" [API SERVER ERROR-{session_id}] Статус {response.status} для '{prompt_input_for_log!r}'. Ответ: {error_text[:500]}")
                            return item, None

                        response_data = await response.json()
                        content = response_data['message']['content'].strip()
                        if content.startswith("```json"): content = content[7:-3].strip()
                        print(f" [API-{session_id}] Вход: '{prompt_input_for_log!r}...' -> Выход: '{content[:100]}...'")
                        await asyncio.sleep(0.5)
                        return item, content

                except aiohttp.ContentTypeError:
                    error_text = await response.text()
                    print(
                        f" [API FORMAT ERROR-{session_id}] Сервер вернул не-JSON для '{prompt_input_for_log!r}'. Ответ: {error_text[:500]}")
                    return item, None
                except Exception as e:
                    print(f" [API UNKNOWN ERROR-{session_id}] Для '{prompt_input_for_log!r}...' ошибка: {repr(e)}")
                    return item, None

        tasks = [get_one(item) for item in items]
        return await asyncio.gather(*tasks)



# Объединяет логику кластеризации и синтеза
async def get_synthesized_groups(remarks_list, all_embeddings, semaphore):
    if not remarks_list:
        return [], {}  # Возвращаем пустые результаты, если нет замечаний

    # Извлекаем эмбеддинги только для текущей группы замечаний
    original_indices = [r['id'] for r in remarks_list]
    current_embeddings = all_embeddings[original_indices]

    remark_clusters = cluster_remarks(remarks_list, current_embeddings, distance_threshold=CLUSTER_DISTANCE_THRESHOLD)

    print(f"\n--- Этап: Подготовка к синтезу (лимит на кластер: {MAX_REMARKS_FOR_SYNTHESIS}) ---")
    items_for_synthesis = []
    for i, cluster in enumerate(remark_clusters):
        if len(cluster) <= 1:
            continue

        cluster_texts = [r['text'] for r in cluster]
        remarks_to_send = cluster_texts

        if len(cluster) > MAX_REMARKS_FOR_SYNTHESIS:
            print(
                f" [CHAMPIONS] Кластер {i} слишком большой ({len(cluster)}). Выбираем {MAX_REMARKS_FOR_SYNTHESIS} чемпионов...")
            cluster_original_indices = [r['id'] for r in cluster]
            cluster_embeddings = all_embeddings[cluster_original_indices]

            centroid = np.mean(cluster_embeddings, axis=0)
            similarities = cosine_similarity(cluster_embeddings, centroid.reshape(1, -1)).flatten()
            top_indices_in_cluster = np.argsort(similarities)[-MAX_REMARKS_FOR_SYNTHESIS:]
            remarks_to_send = [cluster_texts[j] for j in top_indices_in_cluster]

        items_for_synthesis.append({'cluster_id': i, 'texts_list': remarks_to_send})

    cluster_names, synthesized_remarks = {}, {}
    if items_for_synthesis:
        print("\n---  Этап: Запуск синтеза для кластеров ---")
        processing_results = await process_llm_requests(items_for_synthesis, COMPLEX_SYNTHESIS_PROMPT, semaphore,
                                                        session_id='SYNTH-JSON')
        for item, content in processing_results:
            cluster_id = item['cluster_id']
            try:
                data = json.loads(content)
                cluster_names[cluster_id] = data.get('group_name', 'Без названия')
                synthesized_remarks[cluster_id] = data.get('synthesized_remark', item['texts_list'][0])
            except (json.JSONDecodeError, TypeError):
                print(f" [WARNING] Не удалось распарсить JSON для кластера {cluster_id}. Используем запасной вариант.")
                cluster_names[cluster_id] = "Название не сгенерировано"
                synthesized_remarks[cluster_id] = item['texts_list'][0]

    final_groups_list = []
    for i, cluster in enumerate(remark_clusters):
        original_texts = [r['text'] for r in cluster]
        if len(cluster) == 1:
            final_groups_list.append({
                "text_to_classify": original_texts[0],
                "group_name": "Уникальное замечание",
                "original_remarks": original_texts
            })
        else:
            synthesized_text = synthesized_remarks.get(i, original_texts[0])
            group_name = cluster_names.get(i, "Без названия")
            final_groups_list.append({
                "text_to_classify": synthesized_text,
                "group_name": group_name,
                "original_remarks": original_texts
            })

    # Сохраняем отчет о синтезе для отладки
    synthesis_report = [{"group_name": item['group_name'], "synthesized_remark": item['text_to_classify'],
                         "original_duplicates": item['original_remarks']} for item in final_groups_list if
                        len(item['original_remarks']) > 1]

    return final_groups_list, synthesis_report


async def main():
    print("\n---  ЗАПУСК ГИБРИДНОГО ПАЙПЛАЙНА ---")
    semaphore = Semaphore(MAX_CONCURRENT_REQUESTS)

    # === ЭТАП 0: Загрузка и предварительная обработка ===
    major_categories_kb, sub_categories_kb = load_knowledge_base(THEMES_FILE)

    # ### ИЗМЕНЕНИЕ: Используем новую функцию для разделения данных ###
    preclassified_remarks, unclassified_remarks = load_and_partition_remarks(DATA_FILE)

    all_remarks_list = list(itertools.chain.from_iterable(preclassified_remarks.values())) + unclassified_remarks
    if not all_remarks_list:
        print("Не найдено замечаний для обработки.")
        return

    print("\n--- Этап 0.5: Предварительное создание всех эмбеддингов ---")
    # Сортируем по ID, чтобы сохранить порядок для индексации
    all_remarks_list.sort(key=lambda x: x['id'])
    all_remark_texts = [r['text'] for r in all_remarks_list]
    all_remark_embeddings = embedding_model.encode(all_remark_texts, show_progress_bar=True)

    final_report = defaultdict(list)
    synthesis_reports = {}

    # === ЭТАП 1: Обработка НЕКЛАССИФИЦИРОВАННЫХ замечаний (полный цикл) ===
    print("\n\n---  ПАЙПЛАЙН А: Обработка неклассифицированных замечаний (из 'None') ---")
    if unclassified_remarks:
        # Кластеризация и синтез
        unclassified_groups, unclassified_synthesis_report = await get_synthesized_groups(unclassified_remarks,
                                                                                          all_remark_embeddings,
                                                                                          semaphore)
        synthesis_reports['unclassified'] = unclassified_synthesis_report

        # Классификация
        print(f"\n--- Этап: Гибридная классификация для {len(unclassified_groups)} групп ---")
        group_texts_to_classify = [item['text_to_classify'] for item in unclassified_groups]
        group_embeddings = embedding_model.encode(group_texts_to_classify, convert_to_tensor=True, device=device)

        # Получаем эмбеддинги для категорий из базы знаний
        major_cat_embeddings = embedding_model.encode(major_categories_kb, convert_to_tensor=True, device=device)
        sub_cat_embeddings = embedding_model.encode(sub_categories_kb, convert_to_tensor=True, device=device)

        major_cos_scores = util.cos_sim(group_embeddings, major_cat_embeddings)
        major_top_scores, major_top_indices = torch.max(major_cos_scores, dim=1)

        for i, item in enumerate(unclassified_groups):
            text = item['text_to_classify']
            major_cat_score, major_cat_nlp = major_top_scores[i].item(), major_categories_kb[
                major_top_indices[i].item()]

            # --- Определение основной категории ---
            major_cat = "Прочее"
            if major_cat_score >= NLP_CLASSIFICATION_THRESHOLD:
                major_cat = major_cat_nlp
                print(f" [NLP-CLASS] '{text[:50]}...' -> '{major_cat}' (score: {major_cat_score:.2f})")
            else:
                print(
                    f" [NLP-UNSURE] Низкая уверенность для основной категории ({major_cat_score:.2f}). Спрашиваем LLM...")
                major_cat_result, = await process_llm_requests([{'text': text}], CLASSIFY_MAJOR_PROMPT, semaphore,
                                                               categories="\n".join(major_categories_kb),
                                                               session_id='LLM-MAJOR')
                major_cat = major_cat_result[1] if major_cat_result[1] else "Прочее"

            # --- Определение подкатегории ---
            sub_cat = 'Не удалось классифицировать'
            if major_cat != "Прочее":
                # Пересчитываем эмбеддинги для обновленной базы знаний подкатегорий
                sub_cat_embeddings_updated = embedding_model.encode(sub_categories_kb, convert_to_tensor=True,
                                                                    device=device)
                sub_cos_scores = util.cos_sim(group_embeddings[i].unsqueeze(0), sub_cat_embeddings_updated)
                sub_top_score, sub_top_index = torch.max(sub_cos_scores, dim=1)

                if sub_top_score.item() >= NLP_CLASSIFICATION_THRESHOLD:
                    sub_cat = sub_categories_kb[sub_top_index.item()]
                    print(f" [NLP-SUB-CLASS] -> '{sub_cat}' (score: {sub_top_score.item():.2f})")
                else:
                    print(f" [NLP-UNSURE-SUB] Низкая уверенность для подкатегории. Спрашиваем LLM...")
                    sub_cat_result, = await process_llm_requests([{'text': text}], CLASSIFY_SUB_PROMPT, semaphore,
                                                                 sub_categories="\n".join(sub_categories_kb),
                                                                 session_id='LLM-SUB')
                    found_sub_cat = sub_cat_result[1] if sub_cat_result[1] and sub_cat_result[
                        1].lower() != 'none' else None

                    if found_sub_cat:
                        sub_cat = found_sub_cat
                    else:
                        new_sub_cat_result, = await process_llm_requests([{'text': text}], CREATE_NEW_SUB_PROMPT,
                                                                         semaphore, session_id='CREATE-SUB')
                        new_sub_cat = new_sub_cat_result[1]
                        if new_sub_cat:
                            sub_cat = new_sub_cat
                            if new_sub_cat not in sub_categories_kb:
                                sub_categories_kb.append(new_sub_cat)  # Обновляем базу знаний "на лету"

            category_key = f"{major_cat} / {sub_cat}"
            final_report[category_key].append(item)
    else:
        print("Неклассифицированных замечаний не найдено. Пропускаем Пайплайн А.")

    # === ЭТАП 2: Обработка ПРЕДВАРИТЕЛЬНО КЛАССИФИЦИРОВАННЫХ замечаний (упрощенный цикл) ===
    print("\n\n---  ПАЙПЛАЙН Б: Обработка предварительно классифицированных замечаний ---")
    for major_cat, remarks in preclassified_remarks.items():
        print(f"\n--- Обрабатываем категорию: '{major_cat}' ({len(remarks)} шт.) ---")

        # Кластеризация и синтез внутри основной категории
        preclassified_groups, preclassified_synthesis_report = await get_synthesized_groups(remarks,
                                                                                            all_remark_embeddings,
                                                                                            semaphore)
        synthesis_reports[major_cat] = preclassified_synthesis_report

        # Классификация ТОЛЬКО подкатегорий
        print(f"\n--- Этап: Определение подкатегорий для {len(preclassified_groups)} групп в '{major_cat}' ---")
        group_texts_to_classify = [item['text_to_classify'] for item in preclassified_groups]
        group_embeddings = embedding_model.encode(group_texts_to_classify, convert_to_tensor=True, device=device)

        for i, item in enumerate(preclassified_groups):
            text = item['text_to_classify']
            sub_cat = 'Не удалось классифицировать'

            sub_cat_embeddings_updated = embedding_model.encode(sub_categories_kb, convert_to_tensor=True,
                                                                device=device)
            sub_cos_scores = util.cos_sim(group_embeddings[i].unsqueeze(0), sub_cat_embeddings_updated)
            sub_top_score, sub_top_index = torch.max(sub_cos_scores, dim=1)

            if sub_top_score.item() >= NLP_CLASSIFICATION_THRESHOLD:
                sub_cat = sub_categories_kb[sub_top_index.item()]
                print(f" [NLP-SUB-CLASS] '{text[:50]}...' -> '{sub_cat}' (score: {sub_top_score.item():.2f})")
            else:
                print(f" [NLP-UNSURE-SUB] Низкая уверенность для подкатегории. Спрашиваем LLM...")
                sub_cat_result, = await process_llm_requests([{'text': text}], CLASSIFY_SUB_PROMPT, semaphore,
                                                             sub_categories="\n".join(sub_categories_kb),
                                                             session_id='LLM-SUB')
                found_sub_cat = sub_cat_result[1] if sub_cat_result[1] and sub_cat_result[1].lower() != 'none' else None

                if found_sub_cat:
                    sub_cat = found_sub_cat
                else:
                    new_sub_cat_result, = await process_llm_requests([{'text': text}], CREATE_NEW_SUB_PROMPT, semaphore,
                                                                     session_id='CREATE-SUB')
                    new_sub_cat = new_sub_cat_result[1]
                    if new_sub_cat:
                        sub_cat = new_sub_cat
                        if new_sub_cat not in sub_categories_kb:
                            sub_categories_kb.append(new_sub_cat)

            category_key = f"{major_cat} / {sub_cat}"
            final_report[category_key].append(item)

    # === ЭТАП 3: Сборка и сохранение результатов ===
    print("\n\n---  Этап: Сборка и сохранение результатов ---")

    # Обновляем базу знаний новыми подкатегориями, если они появились
    save_knowledge_base(THEMES_FILE, major_categories_kb, sub_categories_kb)

    # Сохраняем отчеты о синтезе
    with open("synthesis_report_clustered.json", "w", encoding="utf-8") as f:
        json.dump(synthesis_reports, f, ensure_ascii=False, indent=2)
    print("  Отчеты о синтезе сохранены в synthesis_report_clustered.json")

    # Формируем и сохраняем финальный отчет
    final_report_list = [{"category": name, "items": items} for name, items in sorted(final_report.items())]
    with open("report_final_classified.json", "w", encoding="utf-8") as f:
        json.dump(final_report_list, f, ensure_ascii=False, indent=2)
    print("  Финальный классифицированный отчет сохранен в report_final_classified.json")
    print("\n---  Пайплайн завершен! ---")


if __name__ == "__main__":
    asyncio.run(main())