let currentShortCode = '';

async function shortenURL() {
    const urlInput = document.getElementById('urlInput');
    const url = urlInput.value.trim();

    if (!url) {
        showError('Please enter a URL');
        return;
    }

    // Hide previous results/errors
    hideAllSections();
    showLoading();

    try {
        const response = await fetch('/api/create', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ url: url }),
        });

        // Check if response is ok before parsing JSON
        if (!response.ok) {
            let errorMessage = 'Failed to shorten URL';
            try {
                const errorData = await response.json();
                errorMessage = errorData.error || errorMessage;
            } catch (e) {
                errorMessage = `Server error: ${response.status} ${response.statusText}`;
            }
            throw new Error(errorMessage);
        }

        const data = await response.json();

        // Store the short code for stats
        currentShortCode = data.short_code;

        // Display result
        document.getElementById('shortURL').value = data.short_url;
        document.getElementById('originalURL').href = data.long_url;
        document.getElementById('originalURL').textContent = data.long_url;

        hideLoading();
        showResult();
    } catch (error) {
        hideLoading();
        // More specific error messages
        if (error.message.includes('Failed to fetch') || error.message.includes('NetworkError')) {
            showError('Cannot connect to server. Please make sure the server is running on http://localhost:8080');
        } else {
            showError(error.message || 'An error occurred. Please try again.');
        }
    }
}

function showResult() {
    hideAllSections();
    document.getElementById('resultSection').classList.remove('hidden');
}

function showStats() {
    if (!currentShortCode) {
        showError('No URL selected. Please shorten a URL first.');
        return;
    }

    hideAllSections();
    showLoading();

    fetch(`/api/stats/${currentShortCode}`)
        .then(async response => {
            // Check if response is ok before parsing JSON
            if (!response.ok) {
                let errorMessage = 'Failed to fetch stats';
                try {
                    const errorData = await response.json();
                    errorMessage = errorData.error || errorMessage;
                } catch (e) {
                    if (response.status === 404) {
                        errorMessage = 'URL not found';
                    } else {
                        errorMessage = `Server error: ${response.status} ${response.statusText}`;
                    }
                }
                throw new Error(errorMessage);
            }
            return response.json();
        })
        .then(data => {
            if (!data) {
                throw new Error('No data received from server');
            }

            document.getElementById('statCode').textContent = data.short_code || currentShortCode;
            document.getElementById('statClicks').textContent = data.click_count || 0;
            
            // Format date
            if (data.created_at) {
                const date = new Date(data.created_at);
                if (!isNaN(date.getTime())) {
                    document.getElementById('statCreated').textContent = date.toLocaleString();
                } else {
                    document.getElementById('statCreated').textContent = data.created_at;
                }
            } else {
                document.getElementById('statCreated').textContent = 'Unknown';
            }
            
            hideLoading();
            document.getElementById('statsSection').classList.remove('hidden');
        })
        .catch(error => {
            hideLoading();
            // More specific error messages
            if (error.message.includes('Failed to fetch') || error.message.includes('NetworkError')) {
                showError('Cannot connect to server. Please make sure the server is running.');
            } else {
                showError(error.message || 'Failed to load statistics');
            }
        });
}

function closeStats() {
    document.getElementById('statsSection').classList.add('hidden');
    showResult();
}

function resetForm() {
    document.getElementById('urlInput').value = '';
    currentShortCode = '';
    hideAllSections();
    document.getElementById('urlInput').focus();
}

function copyToClipboard(elementId) {
    const element = document.getElementById(elementId);
    element.select();
    element.setSelectionRange(0, 99999); // For mobile devices

    try {
        document.execCommand('copy');
        
        // Visual feedback
        const copyBtn = element.nextElementSibling;
        const originalText = copyBtn.querySelector('span').textContent;
        copyBtn.querySelector('span').textContent = 'Copied!';
        copyBtn.style.background = '#059669';
        
        setTimeout(() => {
            copyBtn.querySelector('span').textContent = originalText;
            copyBtn.style.background = '';
        }, 2000);
    } catch (err) {
        showError('Failed to copy to clipboard');
    }
}

function showError(message) {
    hideAllSections();
    document.getElementById('errorMessage').textContent = message;
    document.getElementById('errorSection').classList.remove('hidden');
}

function closeError() {
    document.getElementById('errorSection').classList.add('hidden');
}

function showLoading() {
    document.getElementById('loadingSection').classList.remove('hidden');
}

function hideLoading() {
    document.getElementById('loadingSection').classList.add('hidden');
}

function hideAllSections() {
    document.getElementById('resultSection').classList.add('hidden');
    document.getElementById('statsSection').classList.add('hidden');
    document.getElementById('errorSection').classList.add('hidden');
    document.getElementById('loadingSection').classList.add('hidden');
}

// Allow Enter key to submit
document.getElementById('urlInput').addEventListener('keypress', function(e) {
    if (e.key === 'Enter') {
        shortenURL();
    }
});

// Focus input on page load
window.addEventListener('load', function() {
    const urlInput = document.getElementById('urlInput');
    if (urlInput) urlInput.focus();
});

