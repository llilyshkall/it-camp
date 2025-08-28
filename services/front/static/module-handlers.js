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
    const response = await fetch(endpoints.basicProjects, { 
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
  
/**
 * Отправка файлов из поля загрузки документов чеклиста 
 * @param {Event} event - Событие клика
 * @param {Object} options - Дополнительные параметры
 */
async function sendAssuranceDocuments(event, options = {}) {
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
        `${endpoints.loadFile}${options.projectID}${endpoints.loadFileDocumentation}`, 
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
 * Обработчик кнопки "Начать проверку ашуранса"
 * @param {Event} event - Событие клика
 * @param {Object} options - Дополнительные параметры
 */
async function handleStartAssurance(event, options = {}) {
  console.log('Отправляем на проверку');
  
  const requestData = {
  };

  const saveBtn = document.getElementById('start-assurance');
  if (saveBtn) {
    saveBtn.disabled = true;
    saveBtn.classList.add('loading');
  }

  try {
    const response = await fetch(
      `${endpoints.projects}${options.projectID}${endpoints.startDocCheck}`, 
      { 
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(requestData)
    });

    if (response.status === 404) {
          throw new Error('404 Выполняется обработка. Пожалуйста, попробуйте позже.');
        }
      if (response.status === 409) {
          throw new Error('409 Выполняется обработка. Пожалуйста, попробуйте позже.');
        }

      if (!response.ok) {
          throw new Error(`Ошибка сервера! Статус: ${response.status}`);  
        }

    const data = await response.json();
    console.log('✅ Обработка успешно начата:', data);
    showToast('Обработка успешно начата', true);
    return data;

  } catch (error) {
    console.error('❌ Ошибка при выполнении фунции:', error);
    const errorMessages = {
      '404': 'Выполняется обработка. Пожалуйста, попробуйте позже.',
      '409': 'Выполняется обработка. Пожалуйста, попробуйте позже.',
      'default': 'Не удалось скачать файл'
    };

    const errorMessage = errorMessages[error.message.match(/404|409/)?.[0]] || errorMessages.default;
    showToast(errorMessage, false);
  } finally {
    if (saveBtn) {
      saveBtn.disabled = false;
      saveBtn.classList.remove('loading');
    }
  }
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
      throw new Error('404 Отчёта нет');
    }
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
    
    // Специальное сообщение для 404 ошибки  
    const errorMessages = {
      '404': 'Отчёта нет',
      '409': 'Результат проверки ещё не готов. Пожалуйста, попробуйте позже.',
      'default': 'Не удалось скачать файл'
    };

    const errorMessage = errorMessages[error.message.match(/404|409/)?.[0]] || errorMessages.default;
    
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
async function handleStartRemarks(event, options = {}) {
  console.log('⚙️ Начинаем обработку замечаний...');
  
  const remarksInput = document.getElementById('remarks-input');
  const files = remarksInput.files;
  
  if (!files || files.length === 0) {
    showToast('Выберите файлы с замечаниями', false);
    return;
  }
  
  const startBtn = document.getElementById('start-remarks');
  //const loadingIndicator = document.getElementById('remarks-loading');
  
  startBtn.disabled = true;
  startBtn.classList.add('loading');
  //loadingIndicator.hidden = false;

  try {
    // Последовательно отправляем все файлы
      const file = files[0];
      
      if (!file) {
        console.warn(`Файл не существует`);
        throw new Error(`Ошибка файла ${file.name}: не найден`);
      }
      
      const formData = new FormData();
      formData.append('file', file);
      
      
      const response = await fetch(
        `${endpoints.loadFile}${options.projectID}${endpoints.remarks}`, 
        {
          method: 'POST',
          body: formData
        }
      );
      

        // Обрабатываем 404 ошибку отдельно  
      if (response.status === 404) {
        throw new Error('404 Выполняется обработка. Пожалуйста, попробуйте позже.');
      }
      if (response.status === 409) {
        throw new Error('409 Выполняется обработка. Пожалуйста, попробуйте позже.');
      }

      if (!response.ok) {
        throw new Error(`Ошибка сервера! Статус: ${response.status}`);  
      }
      
      
      const result = await response.json();
      console.log(`✅ Файл ${file.name} успешно обработан:`, result);
    
      
      showToast(`Обработка успешно начата`, true);
      console.log('✅ Обработка успешно начата');  
  } catch (error) {
    console.error('❌ Ошибка при выполнении фунции:', error);
    const errorMessages = {
      '404': 'Выполняется обработка. Пожалуйста, попробуйте позже.',
      '409': 'Выполняется обработка. Пожалуйста, попробуйте позже.',
      'default': 'Не удалось скачать файл'
    };

    const errorMessage = errorMessages[error.message.match(/404|409/)?.[0]] || errorMessages.default;
    showToast(errorMessage, false);
  } finally {
    startBtn.disabled = false;
    startBtn.classList.remove('loading');
  }
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

  //const url = endpoints.projects + options.projectID + "/remarks_clustered";
  const url = `${endpoints.projects}${options.projectID}${endpoints.remarks_clustered}`;
  try {
    const response = await fetch(url, {
      method: 'GET',
      headers: {
        'Accept': 'application/octet-stream'
      }
    });

    // Обрабатываем 404 ошибку отдельно  
    if (response.status === 404) {
      throw new Error('404 Отчёта нет');
    }
    if (response.status === 409) {
      throw new Error('409 Отчёт ещё не готов. Пожалуйста, попробуйте позже.');
    }

    if (!response.ok) {
      throw new Error(`Ошибка сервера! Статус: ${response.status}`);  
    }

    // Получаем имя файла из заголовка Content-Disposition  
    const contentDisposition = response.headers.get('Content-Disposition');
    let filename = 'remarks_report.pdf'; // значение по умолчанию  
    // if (contentDisposition) {
    //   const filenameMatch = contentDisposition.match(/filename="?(.+)"?/);
    //   if (filenameMatch) filename = filenameMatch[1];
    // }

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
    const errorMessages = {
      '404': 'Отчёта нет',
      '409': 'Результат проверки ещё не готов. Пожалуйста, попробуйте позже.',
      'default': 'Не удалось скачать файл'
    };

    const errorMessage = errorMessages[error.message.match(/404|409/)?.[0]] || errorMessages.default;
    
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
async function handleMakeProtocol(event, options = {}) {
  console.log('📄 Формируем протокол...');
  
  const protocolBtn = document.getElementById('make-protocol');
  const loadingIndicator = document.getElementById('protocol-loading');
  
  protocolBtn.disabled = true;
  protocolBtn.classList.add('loading');
  loadingIndicator.hidden = false;

  const requestData = {
  };

  try {
    const response = await fetch(
      `${endpoints.projects}${options.projectID}${endpoints.finalReport}`, 
      { 
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(requestData)
      });

    if (!response.ok) {
      const errorData = await response.json();
      
      // Обработка специфических ошибок
      let errorMessage;
      if (response.status === 404) {
        errorMessage = 'Проект не найден (404)';
      } else if (response.status === 409) {
        errorMessage = 'Формирование протокола уже запущено (409)';
      } else {
        errorMessage = errorData.message || 'Ошибка при создании проекта';
      }
      
      throw new Error(errorMessage);
    }

    const data = await response.json();
    console.log('✅ Формирование протокола успешно начато:', data);
    showToast('Формирование протокола успешно начато', true);
    return data;

  } catch (error) {
    console.error('❌ Ошибка начала формирования протокола:', error);
    
    const errorMessages = {
      '404': 'Не хватает данных для формирования протокола',
      '409': 'Результат проверки ещё не готов. Пожалуйста, попробуйте позже.',
      'default': 'Ошибка при выполнении функции'
    };

    const errorMessage = errorMessages[error.message.match(/404|409/)?.[0]] || errorMessages.default;
    
    showToast(errorMessage, false);
    showToast(userMessage, false);
    throw error;
  } finally {
    if (protocolBtn) {  // исправлено с saveBtn на protocolBtn
      protocolBtn.disabled = false;
      protocolBtn.classList.remove('loading');
      loadingIndicator.hidden = true;
    }
  }
}


/**
 * Обработчик кнопки "Скачать протокол"
 * @param {Event} event - Событие клика
 * @param {Object} options - Дополнительные параметры
 */
async function handleDownloadProtocol(event, options = {}) {
  console.log('📥 Скачиваем финальный отчёт...');
  
  const downloadBtn = document.getElementById('download-protocol');
  if (downloadBtn) {
    downloadBtn.disabled = true;
    downloadBtn.classList.add('loading');
  }

  const url = `${endpoints.projects}${options.projectID}${endpoints.finalReport}`;
  try {
    const response = await fetch(url, {
      method: 'GET',
      headers: {
        'Accept': 'application/octet-stream'
      }
    });

    // Обрабатываем 404 ошибку отдельно  
    if (response.status === 404) {
      throw new Error('404 Отчёта нет');
    }
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
    
    // Специальное сообщение для 404 ошибки  
    const errorMessages = {
      '404': 'Отчёта нет',
      '409': 'Результат проверки ещё не готов. Пожалуйста, попробуйте позже.',
      'default': 'Не удалось скачать файл'
    };

    const errorMessage = errorMessages[error.message.match(/404|409/)?.[0]] || errorMessages.default;
    
    showToast(errorMessage, false);
  } finally {
    if (downloadBtn) {
      downloadBtn.disabled = false;
      downloadBtn.classList.remove('loading');
    }
  }
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

// ===== ЭКСПОРТ ФУНКЦИЙ ДЛЯ ИСПОЛЬЗОВАНИЯ =====

// Делаем функции доступными глобально
window.moduleHandlers = {
  // Модуль проверки по чек-листу
  startAssurance: handleStartAssurance,
  //checkResult: handleCheckResult,
  downloadAssurance: handleDownloadAssurance,
  sendAssuranceDocuments: sendAssuranceDocuments,
  
  // Модуль обработки замечаний
  startRemarks: handleStartRemarks,
  downloadRemarks: handleDownloadRemarks,
  
  // Модуль формирования протокола
  makeProtocol: handleMakeProtocol,
  downloadProtocol: handleDownloadProtocol,
  
  // Вспомогательные функции
  showToast: showToast,
  showAssuranceResult: showAssuranceResult,

  //создать проект
  createProject: createProject
};

console.log('✅ Модуль обработчиков загружен и готов к использованию');
console.log('📋 Доступные обработчики:', Object.keys(window.moduleHandlers));
