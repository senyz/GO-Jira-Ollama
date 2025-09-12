// static/js/scripts.js
let selectedModel = "";

// Инициализация при загрузке
$(document).ready(function() {
    if (window.initialSelectedModel) {
        selectModel(window.initialSelectedModel);
    }
});

function selectModel(modelName) {
    if (!modelName) return;
    
    selectedModel = modelName;
    $('#current-model').html('<i class="fas fa-check"></i> ' + modelName);
    $('#selected-model-count').text(modelName ? 1 : 0);
    $('#ai-submit').prop('disabled', !modelName);
    
    // Устанавливаем выбранное значение в select
    $('#model-select').val(modelName);
    
    // Сохраняем выбор на сервере
    $.post('/select-model', { model: modelName })
        .fail(function() {
            console.error('Ошибка сохранения выбора модели');
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

function getTasks() {
    const projectKey = $('#projectKey').val().trim();
    
    if (!projectKey) {
        alert('Пожалуйста, введите ключ проекта');
        return;
    }

    const btn = $('button').filter(function() {
        return $(this).text().includes('Получить задачи');
    });
    
    btn.prop('disabled', true).html('<span class="loading"></span> Загрузка...');
    
    $.post('/get-tasks', { projectKey: projectKey })
        .done(function(response) {
            if (response.success) {
                updateTasksList(response.tasks);
                $('#tasks-count').text(response.count);
            } else {
                alert('Ошибка при получении задач: ' + (response.error || 'Неизвестная ошибка'));
            }
        })
        .fail(function(xhr) {
            alert('Ошибка сервера: ' + xhr.responseText);
        })
        .always(function() {
            btn.prop('disabled', false).html('<i class="fas fa-search"></i> Получить задачи');
        });
}

function updateTasksList(tasks) {
    let html = '';
    if (tasks && tasks.length > 0) {
        tasks.forEach(task => {
            html += `
                <div class="task-card">
                    <h3>${task.key}: ${task.summary}</h3>
                    <div class="task-info">
                        <p><strong>Статус:</strong> ${task.status}</p>
                        <p><strong>Приоритет:</strong> ${task.priority}</p>
                        <p><strong>Назначена:</strong> ${task.assignee || 'Не назначена'}</p>
                    </div>
                    <p class="task-description">${task.description || 'Описание отсутствует'}</p>
                </div>
            `;
        });
    } else {
        html = '<div class="error">Задачи не найдены</div>';
    }
    $('#tasks-list').html(html);
}