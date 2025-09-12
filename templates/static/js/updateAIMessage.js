function updateAIMessageWithTask() {
    if (!selectedTaskKey) return;
    
    // Находим выбранную задачу
    const taskElement = $(`input[name="selectedTask"][value="${selectedTaskKey}"]`).closest('.task-card');
    const taskSummary = taskElement.find('h3').text();
    const taskDescription = taskElement.find('.task-description').text();
    
    // Добавляем информацию о задаче в сообщение ИИ
    const currentMessage = $('#aiMessages').val();
    const taskInfo = `\n\nЗадача: ${taskSummary}\nОписание: ${taskDescription}`;
    
    $('#aiMessages').val(currentMessage + taskInfo);
}