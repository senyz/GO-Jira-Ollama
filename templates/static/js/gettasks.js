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
    
    // Отправляем данные как JSON
    $.ajax({
        url: '/get-tasks',
        type: 'POST',
        contentType: 'application/json',
        data: JSON.stringify({ projectKey: projectKey }),
        success: function(response) {
            if (response.success) {
                updateTasksList(response.tasks);
                $('#tasks-count').text(response.count);
                // После успешного получения задач фокусируемся на них
                setTimeout(function() {
                    focusOnTasks();
                }, 500);
            } else {
                alert('Ошибка при получении задач: ' + (response.error || 'Неизвестная ошибка'));
            }
        },
        error: function(xhr) {
            let errorMsg = 'Ошибка сервера';
            try {
                const response = JSON.parse(xhr.responseText);
                errorMsg = response.error || xhr.responseText;
            } catch (e) {
                errorMsg = xhr.responseText || 'Неизвестная ошибка';
            }
            alert('Ошибка: ' + errorMsg);
        },
        complete: function() {
            btn.prop('disabled', false).html('<i class="fas fa-search"></i> Получить задачи');
        }
    });
}

function updateTasksList(tasks) {
    console.log('Получены задачи:', tasks);
    
    let html = '';
    if (tasks && tasks.length > 0) {
        tasks.forEach(task => {
            // Правильно извлекаем данные из структуры Jira
            const key = escapeHtml(task.key || '');
            const summary = escapeHtml(task.fields?.summary || '');
            const status = escapeHtml(task.fields?.status?.name || 'Неизвестен');
            const priority = escapeHtml(task.fields?.priority?.name || 'Не указан');
            const assignee = escapeHtml(task.fields?.assignee?.displayName || 'Не назначена');
            const description = escapeHtml(task.fields?.description || 'Описание отсутствует');
            
            html += `
                <div class="task-card">
                    <h3>${key}: ${summary}</h3>
                    <div class="task-info">
                        <p><strong>Статус:</strong> ${status}</p>
                        <p><strong>Приоритет:</strong> ${priority}</p>
                        <p><strong>Назначена:</strong> ${assignee}</p>
                    </div>
                    <p class="task-description">${description}</p>
                    <div class="task-actions">
                        <label class="task-select">
                            <input type="radio" name="selectedTask" value="${key}" 
                                   onchange="selectTaskForAI('${key}')">
                            Выбрать для ИИ
                        </label>
                    </div>
                </div>
            `;
        });
    } else {
        html = '<div class="error">Задачи не найдены</div>';
    }
    $('#tasks-list').html(html);
}

let selectedTaskKey = '';

function selectTaskForAI(taskKey) {
    selectedTaskKey = taskKey;
    console.log('Выбрана задача для ИИ:', taskKey);
    updateAIMessageWithTask();
}