// Функция для применения компактного стиля
function applyCompactStyles() {
    // Уменьшаем заголовки
    $('h2').addClass('compact-header');
    
    // Делаем секции компактнее
    $('.section').addClass('compact-section');
    
    // Обновляем стиль выбранной модели
    $('#selected-model-container').removeClass('selected-model').addClass('compact-selected-model');
    
    console.log('Компактные стили применены');
}

// Функция для фокусировки на задачах
function focusOnTasks() {
    $('body').addClass('tasks-focused');
    $('.tasks-section').addClass('highlighted');
    
    // Плавная прокрутка к задачам
    $('html, body').animate({
        scrollTop: $('.tasks-section').offset().top - 20
    }, 800);
}

// Функция для сброса фокуса
function resetFocus() {
    $('body').removeClass('tasks-focused');
    $('.section').removeClass('highlighted');
}

// Функция для обновления UI выбранной модели (обновленная)
function updateSelectedModelUI(modelName) {
    let container = $('#selected-model-container');
    
    if (container.length === 0) {
        $('.model-selector').append(`
            <div class="compact-selected-model" id="selected-model-container">
                <i class="fas fa-check"></i> Выбрано: <span id="selected-model-text">${modelName}</span>
            </div>
        `);
    } else {
        $('#selected-model-text').text(modelName);
        container.show();
    }
    
    // Применяем компактные стили когда модель выбрана
    applyCompactStyles();
    
    $('#current-model').text(modelName || 'Не выбрана');
    $('#selected-model-count').text(modelName ? 1 : 0);
}

// Функция для обработки успешного получения задач
function handleTasksSuccess() {
    setTimeout(function() {
        focusOnTasks();
    }, 500);
}