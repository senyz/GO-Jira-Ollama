// static/js/uModelsList.js
// Вспомогательная функция для экранирования HTML
function escapeHtml(unsafe) {
    if (!unsafe) return '';
    return unsafe
        .replace(/&/g, "&amp;")
        .replace(/</g, "&lt;")
        .replace(/>/g, "&gt;")
        .replace(/"/g, "&quot;")
        .replace(/'/g, "&#039;");
}
function refreshModels() {
    const btn = $('button').filter(function() {
        return $(this).text().includes('Обновить модели');
    });
    
    btn.prop('disabled', true).html('<span class="loading"></span> Обновление...');
    
    $.get('/api/models')
        .done(function(data) {
            updateModelsList(data);
        })
        .fail(function(xhr) {
            $('#models-list').html('<div class="error">Ошибка загрузки моделей: ' + 
                                  (xhr.responseText || 'Неизвестная ошибка') + '</div>');
            console.error('Ошибка загрузки моделей:', xhr);
        })
        .always(function() {
            btn.prop('disabled', false).html('<i class="fas fa-sync-alt"></i> Обновить модели');
        });
}

function updateModelsList(data) {
    console.log('Получены данные моделей:', data);
    
    let html = '';
    if (data && data.length > 0) {
       /// console.log('Количество моделей:', data.length);
        
        data.forEach(model => {
           // console.log('Обработка модели:', model.name);
            
            // Экранируем все строки для безопасности
            const modelName = escapeHtml(model.name);
            const modelId = escapeHtml(model.model);
            const paramSize = escapeHtml(model.details?.parameter_size || 'N/A');
            const family = escapeHtml(model.details?.family || 'N/A');
            const format = escapeHtml(model.details?.format || 'N/A');
            const quantization = escapeHtml(model.details?.quantization_level || 'N/A');
            
            html += `
                    <div class="model-card">
                        <h3>${modelName}</h3>
                        <div class="model-info">
                            <p><strong>Модель:</strong> ${modelId}</p>
                            <p><strong>Размер:</strong> ${paramSize}</p>
                            <p><strong>Семейство:</strong> ${family}</p>
                            <p><strong>Формат:</strong> ${format}</p>
                            <p><strong>Квантование:</strong> ${quantization}</p>
                        </div>
                        <button class="btn btn-secondary select-model-btn" data-model="${escapeHtml(modelName)}">
                            <i class="fas fa-check"></i> Выбрать
                        </button>
                    </div>
                `;
        });
    } else {
        console.log('Модели не получены или пустой массив');
        html = '<div class="error">Модели не загружены</div>';
    }
    $('#models-list').html(html);
    
    // После обновления списка, установите выбранную модель в select
    if (selectedModel) {
        $('#model-select').val(selectedModel);
    }
}