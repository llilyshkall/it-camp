package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strconv"

	"github.com/xuri/excelize/v2"
)

// Структура для хранения данных
type Model struct {
	ProjectName        string `json:"project_name"`
	ExpertiseDirection string `json:"expertise_direction"`
	ExpertiseSection   string `json:"expertise_section"`
	Text               string `json:"text"`
	Urgency            string `json:"urgency"`
}

type Response struct {
	Status string `json:"status"`
}

// func LoadExcelReestrHandler(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("Content-Type", "application/json")

//		json.NewEncoder(w).Encode(Response{Status: "ok"})
//	}
//
// TODO передавать текст или чтото наподобие
func ParseExcel(filePath string) (string, error) {
	//w.Header().Set("Content-Type", "application/json")
	//filePath := "РЕЕСТР_ягодное - нулевой.xlsx"

	f, err := excelize.OpenFile(filePath)
	if err != nil {
		fmt.Println("Ошибка открытия файла:", err)
		//w.WriteHeader(http.StatusInternalServerError)
		//json.NewEncoder(w).Encode(Response{Status: "int err"})
		return "", err
	}

	sheetName := "Лист1"

	rows, err := f.GetRows(sheetName)
	if err != nil {
		fmt.Println("Ошибка чтения листа:", err)
		//w.WriteHeader(http.StatusInternalServerError)
		//json.NewEncoder(w).Encode(Response{Status: "int err"})
		return "", err
	}

	if len(rows) < 1 {
		fmt.Println("Пустой лист")
		//w.WriteHeader(http.StatusBadRequest)
		//json.NewEncoder(w).Encode(Response{Status: "bad file"})
		return "", err
	}

	headers := rows[0]
	colIndices := map[string]int{}

	requiredCols := []string{
		headers[1], // project_name
		headers[2], // expertise_direction
		headers[3], // expertise_section
		headers[4], // text
		headers[5], // urgency
	}
	//log.Println(requiredCols)
	for i, colName := range headers {
		// switch colName {
		// case "Проект":
		// 	colIndices["project_name"] = i
		// case "Направление экспертизы":
		// 	colIndices["expertise_direction"] = i
		// case "Раздел экспертизы":
		// 	colIndices["expertise_section"] = i
		// case "Содержание рекомендации":
		// 	colIndices["text"] = i
		// case "Срочность":
		// 	colIndices["urgency"] = i
		// }
		switch colName {
		case requiredCols[0]:
			colIndices["project_name"] = i
		case requiredCols[1]:
			colIndices["expertise_direction"] = i
		case requiredCols[2]:
			colIndices["expertise_section"] = i
		case requiredCols[3]:
			colIndices["text"] = i
		case requiredCols[4]:
			colIndices["urgency"] = i
		}
	}

	type TranslationMap map[string]string

	translations := TranslationMap{
		"Программа доизучения (ГРР и ОПР)":                        "reassessment",
		"Сейсмогеологическая модель":                              "seismogeological",
		"Петрофизическая модель":                                  "petrophysical",
		"Геологическая модель":                                    "geological",
		"Разработка и прогноз технологических показателей добычи": "development",
		"Гидродинамическая и интегрированная модели":              "hydrodynamic_integrated",
	}

	// Обработка данных
	var modelList []map[string]string

	for _, row := range rows[1:] { // пропускаем заголовки
		if len(row) < len(requiredCols) {
			continue // пропускаем неполные строки
		}

		projectName := getCell(row, colIndices["project_name"])
		expertiseDirection := getCell(row, colIndices["expertise_direction"])
		expertiseSection := getCell(row, colIndices["expertise_section"])
		text := getCell(row, colIndices["text"])
		urgency := getCell(row, colIndices["urgency"])

		model := map[string]string{
			"project_name":        projectName,
			"expertise_direction": expertiseDirection,
			"expertise_section":   expertiseSection,
			"text":                text,
			"urgency":             urgency,
		}

		modelList = append(modelList, model)
	}
	//log.Println(modelList)
	// Обработка поля expertise_section с проверкой NaN и переводом по словарю
	for i, m := range modelList {
		val := m["expertise_section"]
		if val == "" || val == "None" {
			modelList[i]["expertise_section"] = "None"
			continue
		}

		if numVal, err := parseFloat(val); err == nil && math.IsNaN(numVal) {
			modelList[i]["expertise_section"] = "None"
			continue
		}

		if translated, ok := translations[val]; ok {
			if translated == val {
				modelList[i]["expertise_section"] = "None"
			} else {
				modelList[i]["expertise_section"] = translated
			}
		} else {
			// если нет перевода - оставить как есть или присвоить None по желанию
			if val == "" {
				modelList[i]["expertise_section"] = "None"
			}
		}
	}

	// Группировка по 'expertise_section'
	groupMap := make(map[string][]string)

	for _, m := range modelList {
		key := m["expertise_section"]
		text := m["text"]
		groupMap[key] = append(groupMap[key], text)
	}

	// // Создаем обратную мапу из translations
	// inverseTranslations := make(map[string]string)
	// for k, v := range translations {
	// 	inverseTranslations[v] = k
	// }
	// groupMap["keys"] = inverseTranslations

	// // Сортируем ключи для предсказуемого порядка (опционально)
	// var keys []string
	// for k := range groupMap {
	// 	keys = append(keys, k)
	// }
	// sort.Strings(keys)

	// // Создаем финальный объект для JSON (опционально можно оставить как есть)
	// finalOutput := make(map[string]interface{})
	// for _, k := range keys {
	// 	finalOutput[k] = groupMap[k]
	// }

	// Запись в JSON файл
	outputFile, err := os.Create("data.json")
	if err != nil {
		fmt.Println("Ошибка создания файла:", err)
		return "", err
	}
	defer outputFile.Close()

	// TODO возвращать файл
	encoder := json.NewEncoder(outputFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(groupMap); err != nil {
		fmt.Println("Ошибка записи JSON:", err)
		return "", err
	}
	//json.NewEncoder(w).Encode(Response{Status: "ok"})
	return "", nil
}

// ParseExcelFromBytes парсит Excel файл из байтов и возвращает JSON байты
func ParseExcelFromBytes(fileContent []byte) ([]byte, error) {
	f, err := excelize.OpenReader(bytes.NewReader(fileContent))
	if err != nil {
		return nil, fmt.Errorf("ошибка открытия файла: %w", err)
	}
	defer f.Close()

	sheetName := "Лист1"

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения листа: %w", err)
	}

	if len(rows) < 1 {
		return nil, fmt.Errorf("пустой лист")
	}

	headers := rows[0]
	colIndices := map[string]int{}

	requiredCols := []string{
		headers[1], // project_name
		headers[2], // expertise_direction
		headers[3], // expertise_section
		headers[4], // text
		headers[5], // urgency
	}
	//log.Println(requiredCols)
	for i, colName := range headers {
		switch colName {
		case requiredCols[0]:
			colIndices["project_name"] = i
		case requiredCols[1]:
			colIndices["expertise_direction"] = i
		case requiredCols[2]:
			colIndices["expertise_section"] = i
		case requiredCols[3]:
			colIndices["text"] = i
		case requiredCols[4]:
			colIndices["urgency"] = i
		}
	}

	type TranslationMap map[string]string

	translations := TranslationMap{
		"Программа доизучения (ГРР и ОПР)":                        "reassessment",
		"Сейсмогеологическая модель":                              "seismogeological",
		"Петрофизическая модель":                                  "petrophysical",
		"Геологическая модель":                                    "geological",
		"Разработка и прогноз технологических показателей добычи": "development",
		"Гидродинамическая и интегрированная модели":              "hydrodynamic_integrated",
	}

	// Обработка данных
	var modelList []map[string]string

	for _, row := range rows[1:] { // пропускаем заголовки
		if len(row) < len(requiredCols) {
			continue // пропускаем неполные строки
		}

		projectName := getCell(row, colIndices["project_name"])
		expertiseDirection := getCell(row, colIndices["expertise_direction"])
		expertiseSection := getCell(row, colIndices["expertise_section"])
		text := getCell(row, colIndices["text"])
		urgency := getCell(row, colIndices["urgency"])

		model := map[string]string{
			"project_name":        projectName,
			"expertise_direction": expertiseDirection,
			"expertise_section":   expertiseSection,
			"text":                text,
			"urgency":             urgency,
		}

		modelList = append(modelList, model)
	}
	//log.Println(modelList)

	// Обработка поля expertise_section с проверкой NaN и переводом по словарю
	for i, m := range modelList {
		val := m["expertise_section"]
		if val == "" || val == "None" {
			modelList[i]["expertise_section"] = "None"
			continue
		}

		if numVal, err := parseFloat(val); err == nil && math.IsNaN(numVal) {
			modelList[i]["expertise_section"] = "None"
			continue
		}

		if translated, ok := translations[val]; ok {
			if translated == val {
				modelList[i]["expertise_section"] = "None"
			} else {
				modelList[i]["expertise_section"] = translated
			}
		} else {
			// если нет перевода - оставить как есть или присвоить None по желанию
			if val == "" {
				modelList[i]["expertise_section"] = "None"
			}
		}
	}

	// Группировка по 'expertise_section'
	groupMap := make(map[string][]string)

	for _, m := range modelList {
		key := m["expertise_section"]
		text := m["text"]
		groupMap[key] = append(groupMap[key], text)
	}

	// Преобразуем в JSON байты
	jsonData, err := json.Marshal(groupMap)
	if err != nil {
		return nil, fmt.Errorf("ошибка сериализации в JSON: %w", err)
	}

	return jsonData, nil
}

func getCell(row []string, index int) string {
	if index >= 0 && index < len(row) {
		return row[index]
	}
	return ""
}

func parseFloat(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}
