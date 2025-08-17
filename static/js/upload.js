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

// Initialize all DOM elements and event handlers
document.addEventListener('DOMContentLoaded', function() {
    // File input click handler
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

    // Max LOD slider interaction
    const slider = document.getElementById('max-lod');
    const valueDisplay = document.getElementById('max-lod-value');
    
    if (slider && valueDisplay) {
        slider.addEventListener('input', function() {
            valueDisplay.textContent = this.value;
        });
    }

    // Color field behavior initialization
    const colorInput = document.getElementById('poi-color');
    if (colorInput) {
        // Initially, no color is set
        colorInput.removeAttribute('name');
        ColorFieldManager.setUserSetColor(false);
        ColorFieldManager.updateColorFieldState();
        
        // Track when user changes the color
        colorInput.addEventListener('change', function() {
            ColorFieldManager.setUserSetColor(true);
            colorInput.setAttribute('name', 'poi-color'); // Include in form submission
            ColorFieldManager.updateColorFieldState();
        });
        
        // Also track input events for real-time changes
        colorInput.addEventListener('input', function() {
            if (!ColorFieldManager.getUserSetColor()) {
                ColorFieldManager.setUserSetColor(true);
                colorInput.setAttribute('name', 'poi-color');
                ColorFieldManager.updateColorFieldState();
            }
        });
    }
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


// Color field management module
const ColorFieldManager = (function() {
    let userSetColor = false;
    
    function clearColorField() {
        const colorInput = document.getElementById('poi-color');
        if (colorInput) {
            colorInput.removeAttribute('name'); // Remove from form submission
            colorInput.value = '#000000'; // Reset to default display
            userSetColor = false;
            updateColorFieldState();
        }
    }
    
    function updateColorFieldState() {
        const colorInput = document.getElementById('poi-color');
        const clearBtn = document.querySelector('.clear-color-btn');
        
        if (colorInput && clearBtn) {
            if (userSetColor) {
                colorInput.style.opacity = '1';
                clearBtn.style.display = 'flex';
            } else {
                colorInput.style.opacity = '0.5';
                clearBtn.style.display = 'none';
            }
        }
    }
    
    function setUserSetColor(value) {
        userSetColor = value;
    }
    
    function getUserSetColor() {
        return userSetColor;
    }
    
    // Expose public methods
    return {
        clearColorField: clearColorField,
        updateColorFieldState: updateColorFieldState,
        setUserSetColor: setUserSetColor,
        getUserSetColor: getUserSetColor
    };
})();

// Expose clearColorField globally for onclick handler
function clearColorField() {
    ColorFieldManager.clearColorField();
}

