// static/js/scripts.js
let selectedModel = "";

// Инициализация при загрузке
$(document).ready(function() {
    if (window.initialSelectedModel) {
        selectModel(window.initialSelectedModel);
    }
});

function selectModel(modelName) {
    console.log('Выбор модели:', modelName);
    
    if (!modelName) return;
    
    selectedModel = modelName;
    $('#current-model').html('<i class="fas fa-check"></i> ' + modelName);
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


function sendToAI() {
    if (!selectedModel) {
        alert('Пожалуйста, сначала выберите модель');
        return;
    }

    const messages = $('#aiMessages').val().trim();
    const temperature = $('#temperature').val();
    
    if (!messages) {
        alert('Пожалуйста, введите сообщение');
        return;
    }

    $('#ai-response').removeClass('hidden').html('<div class="loading"></div> ИИ анализирует...');
    
    $.post('/send-to-ai', {
        model: selectedModel,
        messages: messages,
        temperature: temperature
    })
    .done(function(response) {
        $('#ai-response').html(`
            <div class="success">
                <h4><i class="fas fa-robot"></i> Ответ от ${selectedModel}:</h4>
                <p>${response.answer || 'Нет ответа'}</p>
            </div>
        `);
    })
    .fail(function(xhr) {
        $('#ai-response').html(`
            <div class="error">
                <i class="fas fa-exclamation-triangle"></i> Ошибка: ${xhr.responseText || 'Неизвестная ошибка'}
            </div>
        `);
    });
}
