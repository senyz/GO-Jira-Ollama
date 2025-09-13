// static/js/scripts.js
let selectedModel = "";

// Инициализация при загрузке
$(document).ready(function() {
    if (window.initialSelectedModel) {
        selectModel(window.initialSelectedModel);
    }
});

// Функция для выбора модели
function selectModel(modelName) {
    console.log('Выбор модели:', modelName);
    console.log('jQuery доступен:', typeof $ !== 'undefined');
    console.log('Элемент model-selector:', $('.model-selector').length);
    
    if (!modelName) {
        console.log('Модель не выбрана, удаляем контейнер');
        $('#selected-model-container').remove();
        selectedModel = "";
        return;
    }
    
    selectedModel = modelName;
    console.log('Выбрана модель:', selectedModel);
    
    // Обновляем или создаем блок с выбранной моделью
    updateSelectedModelUI(modelName);

    $('#selected-model-count').text(modelName ? 1 : 0);
    $('#ai-submit').prop('disabled', !modelName);
    
    // Устанавливаем выбранное значение в select
    $('#model-select').val(modelName);
    
    // Сохраняем выбор на сервере
    $.post('/select-model', { model: modelName })
        .done(function(response) {
            console.log('Модель сохранена:', response);
        })
        .fail(function(xhr) {
            console.error('Ошибка сохранения выбора модели:', xhr.responseText);
            alert('Ошибка сохранения модели: ' + xhr.responseText);
        });
}

// Функция для обновления UI выбранной модели
function updateSelectedModelUI(modelName) {
    console.log('Обновление UI для модели:', modelName);
    console.log('Контейнер model-selector:', $('.model-selector').length);
    console.log('Текущий контейнер selected-model-container:', $('#selected-model-container').length);
    
    let container = $('#selected-model-container');
    // Создаем контейнер если его нет
    if (container.length === 0) {
    console.log('Создаем новый контейнер');
    $('.model-selector').append(`
        <div class="selected-model" id="selected-model-container" style="display: block !important; background: red !important; color: white !important;">
            <i class="fas fa-check"></i> Выбрано: <span id="selected-model-text">${modelName}</span>
        </div>
    `);

    } else {
        console.log('Обновляем существующий контейнер');
        // Обновляем текст если контейнер существует
        $('#selected-model-text').text(modelName);
        container.show(); // Показываем контейнер
    }
    
    // Обновляем статистику
    $('#current-model').text(modelName || 'Не выбрана');
    $('#selected-model-count').text(modelName ? 1 : 0);
}

function sendToAI() {
    if (!selectedModel) {
        alert('Пожалуйста, сначала выберите модель');
        return;
    }

    const messages = $('#aiMessages').val().trim();
    const temperature = parseFloat($('#temperature').val()) || 0.7;
    
    if (!messages) {
        alert('Пожалуйста, введите сообщение');
        return;
    }

    $('#ai-response').removeClass('hidden').html('<div class="loading"></div> ИИ анализирует...');
    
     const requestData = {
        model: selectedModel,
        messages: messages,
        temperature: temperature,
        taskKey: selectedTaskKey || ''
    };

    console.log('Отправка данных:', requestData);

    $.ajax({
        url: '/send-to-ai',
        type: 'POST',
        contentType: 'application/json',
        data: JSON.stringify(requestData),
        success: function(response) {
            $('#ai-response').html(`
                <div class="success">
                    <h4><i class="fas fa-robot"></i> Ответ от ${selectedModel}:</h4>
                    <p>${response.answer || 'Нет ответа'}</p>
                </div>
            `);
        },
        error: function(xhr) {
            console.error('Ошибка отправки:', xhr);
            $('#ai-response').html(`
                <div class="error">
                    <i class="fas fa-exclamation-triangle"></i> Ошибка: ${xhr.responseText || 'Неизвестная ошибка'}
                </div>
            `);
        }
    });
}