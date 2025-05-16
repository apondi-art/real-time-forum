// login.js - Handles user authentication
document.addEventListener('DOMContentLoaded', function() {
    // Check if user is already logged in
    const token = localStorage.getItem('forum_token');
    if (token) {
        // Redirect to main forum page
        loadMainApplication();
        return;
    }
    renderLoginPage();

    // Use event delegation for dynamic elements
    document.getElementById('app').addEventListener('click', function(e) {
        // Handle register link click
        if (e.target.id === 'registerLink') {
            e.preventDefault();
            renderRegisterPage();
        }
    });

    // Use event delegation for forms too
    document.getElementById('app').addEventListener('submit', function(e) {
        if (e.target.id === 'loginForm') {
            e.preventDefault(); // prevents page reload
            handleLogin();
        }
    });
});

function renderLoginPage() {
    document.getElementById('app').innerHTML = `
        <div class="login-container">
            <h1>Forum Login</h1>
            <form id="loginForm">
                <div class="form-group">
                    <label for="username">Nickname or Email</label>
                    <input type="text" id="username" required>
                </div>
                <div class="form-group">
                    <label for="password">Password</label>
                    <input type="password" id="password" required>
                </div>
                <button type="submit" class="btn">Login</button>
                <div id="errorMessage" class="error-message" style="display: none;"></div>
            </form>
            <div class="links">
                <a href="#" id="registerLink">Create account</a>
            </div>
        </div>
    `;
}

function handleLogin() {
    const username = document.getElementById('username').value;
    const password = document.getElementById('password').value;

    if (!username || !password) {
        showError('Please fill in all fields');
        return;
    }

    // Login API call
    fetch('/api/login', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({
            identifier: username, // Nickname or email
            password: password
        })
    })
    .then(response => {
        if (!response.ok) {
            return response.json().then(data => {
                throw new Error(data.message || 'Login failed');
            });
        }
        return response.json();
    })
    .then(data => {
        // Store token in localStorage
        localStorage.setItem('forum_token', data.token);
        localStorage.setItem('user_id', data.userId);
        localStorage.setItem('username', data.username);

        // Load main application
        loadMainApplication();
    })
    .catch(error => {
        showError(error.message || 'Login failed. Please try again.');
    });
}

function renderRegisterPage() {
    // Check if already loaded
    if (typeof window.renderRegistrationForm === 'function') {
        window.renderRegistrationForm();
        return;
    }

    // Load register.js if not already loaded
    const script = document.createElement('script');
    script.src = '/static/js/register.js'; // Fixed path
    script.onload = function() {
        // Now that the script is loaded, call the function
        if (typeof window.renderRegistrationForm === 'function') {
            window.renderRegistrationForm();
        } else {
            showError('Error: renderRegistrationForm not found after loading.');
            console.error('renderRegistrationForm function not found immediately after script load');
        }
    };
    script.onerror = function() {
        showError('Failed to load registration form');
        console.error('Failed to load register.js script');
    };
    document.head.appendChild(script);
}

function loadMainApplication() {
    // Load main application script
    const mainScript = document.createElement('script');
    mainScript.src = '/static/js/main.js'; // Fixed path
    mainScript.onload = function() {
        console.log('Main application loaded successfully');
        // You might want to call an initialization function here
    };
    mainScript.onerror = function() {
        showError('Failed to load the application');
        console.error('Failed to load main.js script');
    };
    document.head.appendChild(mainScript);
}

function showError(message) {
    const errorElement = document.getElementById('errorMessage');
    if (errorElement) {
        errorElement.textContent = message;
        errorElement.style.display = 'block';
    } else {
        console.error('Error element not found:', message);
    }
}