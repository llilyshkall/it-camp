/**
 * Обработчики для всех кнопок в модулях
 * Готовы для подключения к backend API
 */
import { endpoints } from './config.js';
// ===== МОДУЛЬ 1: Проверка по чек-листу =====


/**
 * Обработчик кнопки "Сохранить в попапе добавления проекта"
 * @param {String} name - Имя проекта
 * @param {String} desc - Описание проекта
 */
async function createProject(name, desc) {
  console.log('🔄 Создание проекта...');
  
  const requestData = {
    name: name,
    description: desc
  };

  const saveBtn = document.getElementById('save-project');
  if (saveBtn) {
    saveBtn.disabled = true;
    saveBtn.classList.add('loading');
  }

  try {
    const response = await fetch(endpoints.createProject, { 
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(requestData)
    });

    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.message || 'Ошибка при создании проекта');
    }

    const data = await response.json();
    console.log('✅ Проект создан:', data);
    showToast('Проект успешно создан', true);
    return data;

  } catch (error) {
    console.error('❌ Ошибка создания проекта:', error);
    showToast(error.message || 'Не удалось создать проект', false);
    throw error;
  } finally {
    if (saveBtn) {
      saveBtn.disabled = false;
      saveBtn.classList.remove('loading');
    }
  }
}
  



async function handleStartAssurance(event, options = {}) {
  console.log('🔄 Начинаем проверку по чек-листу...');
  
  const fileInput = document.getElementById('file-input');
  const files = fileInput.files;
  
  if (!files || files.length === 0) {
    showToast('Выберите файлы для проверки', false);
    return;
  }
  
  const loadingIndicator = document.getElementById('assurance-loading');
  const startBtn = document.getElementById('start-assurance');
  const progressBar = document.getElementById('upload-progress'); // Добавьте элемент прогресса
  
  startBtn.disabled = true;
  startBtn.classList.add('loading');
  loadingIndicator.hidden = false;
  if (progressBar) {
    progressBar.max = files.length;
    progressBar.value = 0;
    progressBar.hidden = false;
  }

  try {
    // Последовательно отправляем все файлы
    for (let i = 0; i < files.length; i++) {
      const file = files[i];
      
      if (!file) {
        console.warn(`Файл ${i} не существует`);
        continue;
      }
      
      const formData = new FormData();
      formData.append('file', file);
      
      // Обновляем статус загрузки
      if (progressBar) {
        progressBar.value = i;
        progressBar.textContent = `${i+1}/${files.length} ${file.name}`;
      }
      
      console.log(`📤 Отправка файла ${i+1}/${files.length}: ${file.name}`);
      showToast(`Отправка файла ${i+1}/${files.length}...`, true);
      
      const response = await fetch(
        `${endpoints.loadFile}${options.projectID}/files?type=documentation`, 
        {
          method: 'POST',
          body: formData
        }
      );
      
      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(`Ошибка файла ${file.name}: ${errorData.message || response.statusText}`);
      }
      
      const result = await response.json();
      console.log(`✅ Файл ${file.name} успешно обработан:`, result);
    }
    
    showToast(`Все файлы (${files.length}) успешно отправлены`, true);
    console.log('✅ Все файлы успешно обработаны');
    
    // Активируем кнопки для следующих действий
    document.getElementById('check-result').disabled = false;
    document.getElementById('download-assurance').disabled = false;
    
  } catch (error) {
    console.error('❌ Ошибка при отправке файлов:', error);
    showToast(error.message || 'Ошибка при отправке файлов', false);
    throw error;
    
  } finally {
    startBtn.disabled = false;
    startBtn.classList.remove('loading');
    loadingIndicator.hidden = true;
    if (progressBar) {
      progressBar.hidden = true;
    }
  }
}

/**
 * Обработчик кнопки "Проверить результат"
 * @param {Event} event - Событие клика
 * @param {Object} options - Дополнительные параметры
 */
function handleCheckResult(event, options = {}) {
  console.log('🔍 Проверяем результат...');
  
  const checkBtn = document.getElementById('check-result');
  checkBtn.disabled = true;
  checkBtn.classList.add('loading');
  
  // TODO: Здесь будет вызов backend API для получения результата
  const requestData = {
    action: 'check_assurance_result',
    projectId: window.currentProject?.id,
    timestamp: new Date().toISOString(),
    ...options
  };
  
  console.log('📤 Запрос результата проверки:', requestData);
  
  // Имитация запроса результата
  simulateBackendCall('/api/assurance/result', requestData)
    .then(response => {
      console.log('✅ Результат проверки:', response);
      
      // Показываем результат
      showAssuranceResult(response);
      showToast('Результат получен');
    })
    .catch(error => {
      console.error('❌ Ошибка при получении результата:', error);
      showToast('Ошибка при получении результата', false);
    })
    .finally(() => {
      checkBtn.disabled = false;
      checkBtn.classList.remove('loading');
    });
}

/**
 * Обработчик кнопки "Скачать отчёт"
 * @param {Event} event - Событие клика
 * @param {Object} options - Дополнительные параметры
 */
async function handleDownloadAssurance(event, options = {}) {
  console.log('📥 Скачиваем отчёт...');
  
  const downloadBtn = document.getElementById('download-assurance');
  if (downloadBtn) {
    downloadBtn.disabled = true;
    downloadBtn.classList.add('loading');
  }

  const url = endpoints.projects + options.projectID + "/checklist";
  try {
    const response = await fetch(url, {
      method: 'GET',
      headers: {
        'Accept': 'application/octet-stream'
      }
    });

    // Обрабатываем 404 ошибку отдельно  
    if (response.status === 404) {
      throw new Error('404 Отчёт ещё не готов. Пожалуйста, попробуйте позже.');
    }

    if (!response.ok) {
      throw new Error(`Ошибка сервера! Статус: ${response.status}`);  
    }

    // Получаем имя файла из заголовка Content-Disposition  
    const contentDisposition = response.headers.get('Content-Disposition');
    let filename = 'report.xlsx'; // значение по умолчанию  
    if (contentDisposition) {
      const filenameMatch = contentDisposition.match(/filename="?(.+)"?/);
      if (filenameMatch) filename = filenameMatch[1];
    }

    // Получаем blob  
    const blob = await response.blob();
    
    // Создаем ссылку для скачивания  
    const downloadUrl = window.URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = downloadUrl;
    link.download = filename;
    link.style.display = 'none';
    document.body.appendChild(link);
    
    // Запускаем скачивание  
    link.click();
    
    // Очищаем  
    setTimeout(() => {
      document.body.removeChild(link);
      window.URL.revokeObjectURL(downloadUrl);
    }, 100);

    console.log('✅ Файл успешно скачан');
    showToast('Файл успешно скачан', true);

  } catch (error) {
    console.error('❌ Ошибка при скачивании файла:', error);
    
    // Специальное сообщение для 404 ошибки  
    const errorMessage = error.message.includes('404') 
      ? 'Результат проверки ещё не готов. Пожалуйста, попробуйте позже.' 
      : 'Не удалось скачать файл';
    
    showToast(errorMessage, false);
  } finally {
    if (downloadBtn) {
      downloadBtn.disabled = false;
      downloadBtn.classList.remove('loading');
    }
  }
}
// ===== МОДУЛЬ 2: Обработка замечаний =====

/**
 * Обработчик кнопки "Обработать замечания"
 * @param {Event} event - Событие клика
 * @param {Object} options - Дополнительные параметры
 */
function handleStartRemarks(event, options = {}) {
  console.log('⚙️ Начинаем обработку замечаний...');
  
  const remarksInput = document.getElementById('remarks-input');
  const files = remarksInput.files;
  
  if (!files || files.length === 0) {
    showToast('Выберите файлы с замечаниями', false);
    return;
  }
  
  const startBtn = document.getElementById('start-remarks');
  const loadingIndicator = document.getElementById('remarks-loading');
  
  startBtn.disabled = true;
  startBtn.classList.add('loading');
  loadingIndicator.hidden = false;
  
  // TODO: Здесь будет вызов backend API для обработки замечаний
  const requestData = {
    action: 'process_remarks',
    projectId: window.currentProject?.id,
    files: Array.from(files).map(f => ({
      name: f.name,
      size: f.size,
      type: f.type
    })),
    processingOptions: {
      normalizeExcel: true,
      generateRegistry: true,
      outputFormat: 'xlsx'
    },
    timestamp: new Date().toISOString(),
    ...options
  };
  
  console.log('📤 Данные для обработки замечаний:', requestData);
  
  // Имитация обработки
  simulateBackendCall('/api/remarks/process', requestData)
    .then(response => {
      console.log('✅ Замечания обработаны:', response);
      
      // Активируем кнопку скачивания
      document.getElementById('download-remarks').disabled = false;
      showToast(`Обработано ${files.length} файлов`);
    })
    .catch(error => {
      console.error('❌ Ошибка при обработке замечаний:', error);
      showToast('Ошибка при обработке замечаний', false);
    })
    .finally(() => {
      startBtn.disabled = false;
      startBtn.classList.remove('loading');
      loadingIndicator.hidden = true;
    });
}

/**
 * Обработчик кнопки "Скачать реестр"
 * @param {Event} event - Событие клика
 * @param {Object} options - Дополнительные параметры
 */
async function handleDownloadRemarks(event, options = {}) {
  console.log('📥 Скачиваем реестр замечаний...');
  
  const downloadBtn = document.getElementById('download-remarks');
  if (downloadBtn) {
    downloadBtn.disabled = true;
    downloadBtn.classList.add('loading');
  }

  const url = endpoints.projects + options.projectID + "/remarks_clustered";
  try {
    const response = await fetch(url, {
      method: 'GET',
      headers: {
        'Accept': 'application/octet-stream'
      }
    });

    // Обрабатываем 409 ошибку отдельно  
    if (response.status === 409) {
      throw new Error('409 Отчёт ещё не готов. Пожалуйста, попробуйте позже.');
    }

    if (!response.ok) {
      throw new Error(`Ошибка сервера! Статус: ${response.status}`);  
    }

    // Получаем имя файла из заголовка Content-Disposition  
    const contentDisposition = response.headers.get('Content-Disposition');
    let filename = 'report.xlsx'; // значение по умолчанию  
    if (contentDisposition) {
      const filenameMatch = contentDisposition.match(/filename="?(.+)"?/);
      if (filenameMatch) filename = filenameMatch[1];
    }

    // Получаем blob  
    const blob = await response.blob();
    
    // Создаем ссылку для скачивания  
    const downloadUrl = window.URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = downloadUrl;
    link.download = filename;
    link.style.display = 'none';
    document.body.appendChild(link);
    
    // Запускаем скачивание  
    link.click();
    
    // Очищаем  
    setTimeout(() => {
      document.body.removeChild(link);
      window.URL.revokeObjectURL(downloadUrl);
    }, 100);

    console.log('✅ Файл успешно скачан');
    showToast('Файл успешно скачан', true);

  } catch (error) {
    console.error('❌ Ошибка при скачивании файла:', error);
    
    // Специальное сообщение для 409 ошибки  
    const errorMessage = error.message.includes('409') 
      ? 'Результат обработки ещё не готов. Пожалуйста, попробуйте позже.' 
      : 'Не удалось скачать файл';
    
    showToast(errorMessage, false);
  } finally {
    if (downloadBtn) {
      downloadBtn.disabled = false;
      downloadBtn.classList.remove('loading');
    }
  }
}

// ===== МОДУЛЬ 3: Формирование протокола =====

/**
 * Обработчик кнопки "Сформировать протокол"
 * @param {Event} event - Событие клика
 * @param {Object} options - Дополнительные параметры
 */
function handleMakeProtocol(event, options = {}) {
  console.log('📄 Формируем протокол...');
  
  const protocolBtn = document.getElementById('make-protocol');
  const loadingIndicator = document.getElementById('protocol-loading');
  
  protocolBtn.disabled = true;
  protocolBtn.classList.add('loading');
  loadingIndicator.hidden = false;
  
  // TODO: Здесь будет вызов backend API для формирования протокола
  const requestData = {
    action: 'generate_protocol',
    projectId: window.currentProject?.id,
    projectName: window.currentProject?.name,
    protocolOptions: {
      format: options.format || 'docx', // docx, pdf
      includeAttachments: true,
      template: options.template || 'default',
      language: options.language || 'ru'
    },
    timestamp: new Date().toISOString(),
    ...options
  };
  
  console.log('📤 Данные для формирования протокола:', requestData);
  
  // Имитация формирования протокола
  simulateBackendCall('/api/protocol/generate', requestData)
    .then(response => {
      console.log('✅ Протокол сформирован:', response);
      
      if (response.downloadUrl) {
        const link = document.createElement('a');
        link.href = response.downloadUrl;
        link.download = response.filename || 'protocol.docx';
        document.body.appendChild(link);
        link.click();
        link.remove();
        showToast('Протокол готов и скачан');
      }
    })
    .catch(error => {
      console.error('❌ Ошибка при формировании протокола:', error);
      showToast('Ошибка при формировании протокола', false);
    })
    .finally(() => {
      protocolBtn.disabled = false;
      protocolBtn.classList.remove('loading');
      loadingIndicator.hidden = true;
    });
}

// ===== ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ =====

/**
 * Показывает результат проверки ашуренса
 * @param {Object} result - Результат проверки
 */
function showAssuranceResult(result) {
  const resultElement = document.getElementById('assurance-result');
  const badge = document.getElementById('verdict-badge');
  const reasonsList = document.getElementById('verdict-reasons');
  
  if (resultElement && badge && reasonsList) {
    badge.textContent = result.title || '—';
    badge.className = 'badge ' + (result.verdict === 'ok' ? 'ok' : 'fail');
    
    reasonsList.innerHTML = '';
    if (result.reasons && Array.isArray(result.reasons)) {
      result.reasons.forEach(reason => {
        const li = document.createElement('li');
        li.textContent = reason;
        reasonsList.appendChild(li);
      });
    }
    
    resultElement.hidden = false;
  }
}

/**
 * Показывает toast сообщение
 * @param {string} message - Текст сообщения
 * @param {boolean} isSuccess - Успешное ли сообщение
 */
function showToast(message, isSuccess = true) {
  const toast = document.getElementById('toast');
  const toastMsg = document.getElementById('toast-msg');
  
  if (toast && toastMsg) {
    toastMsg.textContent = message;
    toast.classList.toggle('bad', !isSuccess);
    toast.hidden = false;
    toast.classList.add('show');
    
    setTimeout(() => {
      toast.classList.remove('show');
    }, 2400);
  }
}

/**
 * Имитирует вызов backend API (заменить на реальные fetch запросы)
 * @param {string} url - URL API endpoint
 * @param {Object} data - Данные для отправки
 * @returns {Promise} - Promise с ответом
 */
function simulateBackendCall(url, data) {
  return new Promise((resolve, reject) => {
    // Имитируем задержку сети
    setTimeout(() => {
      // Имитируем успешный ответ
      if (Math.random() > 0.1) { // 90% успеха
        const response = {
          success: true,
          message: 'Операция выполнена успешно',
          timestamp: new Date().toISOString(),
          data: {
            id: Math.random().toString(36).substr(2, 9),
            status: 'completed'
          }
        };
        
        // Добавляем специфичные данные в зависимости от действия
        if (data.action === 'start_assurance') {
          response.data.assuranceId = 'ass_' + Math.random().toString(36).substr(2, 9);
          response.data.estimatedTime = '2-3 минуты';
        } else if (data.action === 'check_assurance_result') {
          response.data.verdict = 'ok';
          response.data.title = 'Готов к ашурансу';
          response.data.reasons = ['Все файлы соответствуют требованиям', 'Размер в пределах лимита'];
        } else if (data.action === 'download_assurance_report') {
          response.data.downloadUrl = 'data:text/plain;base64,';
          response.data.filename = 'assurance_report.xlsx';
        } else if (data.action === 'process_remarks') {
          response.data.processedFiles = data.files.length;
          response.data.registryId = 'reg_' + Math.random().toString(36).substr(2, 9);
        } else if (data.action === 'download_remarks_registry') {
          response.data.downloadUrl = 'data:text/plain;base64,';
          response.data.filename = 'remarks_registry.xlsx';
        } else if (data.action === 'generate_protocol') {
          response.data.protocolId = 'prot_' + Math.random().toString(36).substr(2, 9);
          response.data.downloadUrl = 'data:text/plain;base64,';
          response.data.filename = 'protocol.docx';
        }
        
        resolve(response);
      } else {
        // Имитируем ошибку
        reject(new Error('Симулированная ошибка backend API'));
      }
    }, 1500 + Math.random() * 1000); // Задержка 1.5-2.5 секунды
  });
}

// ===== ЭКСПОРТ ФУНКЦИЙ ДЛЯ ИСПОЛЬЗОВАНИЯ =====

// Делаем функции доступными глобально
window.moduleHandlers = {
  // Модуль проверки по чек-листу
  startAssurance: handleStartAssurance,
  checkResult: handleCheckResult,
  downloadAssurance: handleDownloadAssurance,
  
  // Модуль обработки замечаний
  startRemarks: handleStartRemarks,
  downloadRemarks: handleDownloadRemarks,
  
  // Модуль формирования протокола
  makeProtocol: handleMakeProtocol,
  
  // Вспомогательные функции
  showToast: showToast,
  showAssuranceResult: showAssuranceResult,

  //создать проект
  createProject: createProject
};

console.log('✅ Модуль обработчиков загружен и готов к использованию');
console.log('📋 Доступные обработчики:', Object.keys(window.moduleHandlers));
