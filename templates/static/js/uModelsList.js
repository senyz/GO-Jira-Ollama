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

function updateModelsList(data) {
    let html = '';
    if (data && data.length > 0) {
        data.forEach(model => {
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
                    <button class="btn btn-secondary" onclick="selectModel('${modelName.replace(/'/g, "\\'")}')">
                        <i class="fas fa-check"></i> Выбрать
                    </button>
                </div>
            `;
        });
    } else {
        html = '<div class="error">Модели не загружены</div>';
    }
    $('#models-list').html(html);
}