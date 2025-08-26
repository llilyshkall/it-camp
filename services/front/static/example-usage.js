/**
 * Примеры использования обработчиков модулей
 * Этот файл демонстрирует различные способы вызова обработчиков
 */

// ===== ПРИМЕР 1: Базовое использование =====

// Простой вызов без параметров
function basicUsage() {
  console.log('=== Базовое использование ===');
  
  // Проверка по чек-листу
  window.moduleHandlers.startAssurance();
  window.moduleHandlers.checkResult();
  window.moduleHandlers.downloadAssurance();
  
  // Обработка замечаний
  window.moduleHandlers.startRemarks();
  window.moduleHandlers.downloadRemarks();
  
  // Формирование протокола
  window.moduleHandlers.makeProtocol();
}

// ===== ПРИМЕР 2: С дополнительными параметрами =====

function advancedUsage() {
  console.log('=== Расширенное использование ===');
  
  // Запуск проверки с высоким приоритетом
  window.moduleHandlers.startAssurance(null, {
    priority: 'high',
    notifyOnComplete: true,
    customOption: 'value'
  });
  
  // Скачивание отчёта в PDF формате
  window.moduleHandlers.downloadAssurance(null, {
    format: 'pdf',
    includeChecklist: true,
    includeMetadata: true
  });
  
  // Обработка замечаний с настройками
  window.moduleHandlers.startRemarks(null, {
    normalizeExcel: true,
    generateRegistry: true,
    outputFormat: 'xlsx'
  });
  
  // Формирование протокола с шаблоном
  window.moduleHandlers.makeProtocol(null, {
    format: 'pdf',
    template: 'custom',
    language: 'en',
    includeAttachments: false
  });
}

// ===== ПРИМЕР 3: Программное использование =====

function programmaticUsage() {
  console.log('=== Программное использование ===');
  
  // Проверяем доступность обработчиков
  if (window.moduleHandlers) {
    console.log('✅ Обработчики доступны');
    console.log('📋 Доступные функции:', Object.keys(window.moduleHandlers));
  } else {
    console.error('❌ Обработчики не загружены');
    return;
  }
  
  // Проверяем текущий проект
  if (window.currentProject) {
    console.log('📁 Текущий проект:', window.currentProject.name);
  } else {
    console.warn('⚠️ Проект не выбран');
  }
  
  // Вызываем обработчики программно
  const options = {
    testMode: true,
    mockResponse: { success: true, data: { test: 'data' } }
  };
  
  // Можно вызывать без события
  window.moduleHandlers.startAssurance(null, options);
}

// ===== ПРИМЕР 4: Обработка событий =====

function eventHandling() {
  console.log('=== Обработка событий ===');
  
  // Добавляем обработчики на кнопки программно
  const startButton = document.getElementById('start-assurance');
  if (startButton) {
    startButton.addEventListener('click', (event) => {
      console.log('🖱️ Клик по кнопке "Начать проверку"');
      
      // Вызываем обработчик с дополнительными параметрами
      window.moduleHandlers.startAssurance(event, {
        priority: 'normal',
        notifyOnComplete: true,
        source: 'button_click'
      });
    });
  }
  
  // Обработчик для кнопки проверки результата
  const checkButton = document.getElementById('check-result');
  if (checkButton) {
    checkButton.addEventListener('click', (event) => {
      console.log('🔍 Проверяем результат...');
      
      window.moduleHandlers.checkResult(event, {
        includeDetails: true,
        format: 'detailed',
        refreshCache: false
      });
    });
  }
}

// ===== ПРИМЕР 5: Кастомные обработчики =====

function customHandlers() {
  console.log('=== Кастомные обработчики ===');
  
  // Создаём кастомный обработчик
  function customAssuranceHandler(event, options = {}) {
    console.log('🎯 Кастомный обработчик ашуренса');
    
    // Предварительная валидация
    if (!window.currentProject) {
      window.moduleHandlers.showToast('Выберите проект', false);
      return;
    }
    
    // Логирование
    console.log('📊 Параметры:', options);
    console.log('📁 Проект:', window.currentProject);
    
    // Вызываем стандартный обработчик с дополнительными параметрами
    const enhancedOptions = {
      ...options,
      customHandler: true,
      timestamp: new Date().toISOString(),
      sessionId: Math.random().toString(36).substr(2, 9)
    };
    
    window.moduleHandlers.startAssurance(event, enhancedOptions);
  }
  
  // Добавляем в глобальный объект
  window.moduleHandlers.customAssurance = customAssuranceHandler;
  
  console.log('✅ Кастомный обработчик добавлен');
}

// ===== ПРИМЕР 6: Тестирование и отладка =====

function testingAndDebug() {
  console.log('=== Тестирование и отладка ===');
  
  // Тестируем обработчики без реального API
  const testOptions = {
    testMode: true,
    mockResponse: {
      success: true,
      data: {
        testId: 'test_' + Math.random().toString(36).substr(2, 9),
        status: 'completed',
        message: 'Тестовый ответ'
      }
    }
  };
  
  // Тестируем все обработчики
  console.log('🧪 Тестируем обработчики...');
  
  try {
    window.moduleHandlers.startAssurance(null, testOptions);
    window.moduleHandlers.checkResult(null, testOptions);
    window.moduleHandlers.downloadAssurance(null, testOptions);
    window.moduleHandlers.startRemarks(null, testOptions);
    window.moduleHandlers.downloadRemarks(null, testOptions);
    window.moduleHandlers.makeProtocol(null, testOptions);
    
    console.log('✅ Все тесты прошли успешно');
  } catch (error) {
    console.error('❌ Ошибка при тестировании:', error);
  }
}

// ===== ИНИЦИАЛИЗАЦИЯ ПРИМЕРОВ =====

// Функция для запуска всех примеров
function runAllExamples() {
  console.log('🚀 Запуск примеров использования обработчиков модулей');
  console.log('==================================================');
  
  // Ждём загрузки DOM
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => {
      setTimeout(runExamples, 1000); // Даём время на загрузку обработчиков
    });
  } else {
    setTimeout(runExamples, 1000);
  }
}

function runExamples() {
  try {
    basicUsage();
    setTimeout(advancedUsage, 500);
    setTimeout(programmaticUsage, 1000);
    setTimeout(eventHandling, 1500);
    setTimeout(customHandlers, 2000);
    setTimeout(testingAndDebug, 2500);
    
    console.log('🎉 Все примеры выполнены!');
  } catch (error) {
    console.error('❌ Ошибка при выполнении примеров:', error);
  }
}

// Автозапуск при загрузке файла
if (typeof window !== 'undefined') {
  // Запускаем примеры через 2 секунды после загрузки страницы
  setTimeout(runAllExamples, 2000);
}

// Экспорт функций для использования в других файлах
if (typeof module !== 'undefined' && module.exports) {
  module.exports = {
    basicUsage,
    advancedUsage,
    programmaticUsage,
    eventHandling,
    customHandlers,
    testingAndDebug,
    runAllExamples
  };
}
