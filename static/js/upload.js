// File upload handling
function handleDragOver(e) {
    e.preventDefault();
    e.target.closest('.upload-area').classList.add('dragover');
}

function handleDragLeave(e) {
    e.target.closest('.upload-area').classList.remove('dragover');
}

function handleDrop(e) {
    e.preventDefault();
    const uploadArea = e.target.closest('.upload-area');
    uploadArea.classList.remove('dragover');
    
    const files = e.dataTransfer.files;
    const fileInput = document.getElementById('file-input');
    fileInput.files = files;
    updateFileList();
}

function updateFileList() {
    const fileInput = document.getElementById('file-input');
    const fileList = document.getElementById('file-list');
    const submitButton = document.getElementById('submit-button');
    const buttonText = document.getElementById('button-text');
    const files = Array.from(fileInput.files);
    
    if (files.length > 0) {
        fileList.innerHTML = '<strong>Selected files:</strong><br>' + 
            files.map(f => `• ${f.name} (${(f.size / 1024 / 1024).toFixed(2)} MB)`).join('<br>');
        if (submitButton && buttonText) {
            submitButton.disabled = false;
            buttonText.textContent = 'Convert to NIMBY Rails Mod';
        }
    } else {
        fileList.innerHTML = '';
        if (submitButton && buttonText) {
            submitButton.disabled = true;
            buttonText.textContent = 'Please select files first';
        }
    }
}

// Initialize file input click handler
document.addEventListener('DOMContentLoaded', function() {
    const uploadClickArea = document.querySelector('.upload-click-area');
    if (uploadClickArea) {
        uploadClickArea.addEventListener('click', function() {
            document.getElementById('file-input').click();
        });
    }

    // Add change event listener to file input
    const fileInput = document.getElementById('file-input');
    if (fileInput) {
        fileInput.addEventListener('change', updateFileList);
    }

    // Set up drag and drop event listeners
    const uploadArea = document.querySelector('.upload-area');
    if (uploadArea) {
        uploadArea.addEventListener('dragover', handleDragOver);
        uploadArea.addEventListener('dragleave', handleDragLeave);
        uploadArea.addEventListener('drop', handleDrop);
    }

    // Initialize button state
    updateFileList();
});

// HTMX event handling for loading spinners and form validation
document.body.addEventListener('htmx:beforeRequest', function(e) {
    // Validate file selection before allowing request
    if (e.target.tagName === 'FORM') {
        const fileInput = e.target.querySelector('#file-input');
        if (fileInput && fileInput.files.length === 0) {
            e.preventDefault();
            alert('Please select at least one file to upload.');
            return false;
        }
    }
    
    const indicator = e.target.getAttribute('hx-indicator');
    if (indicator) {
        const indicatorEl = document.querySelector(indicator);
        if (indicatorEl) indicatorEl.classList.remove('hidden');
    }
});

document.body.addEventListener('htmx:afterRequest', function(e) {
    const indicator = e.target.getAttribute('hx-indicator');
    if (indicator) {
        const indicatorEl = document.querySelector(indicator);
        if (indicatorEl) indicatorEl.classList.add('hidden');
    }
});