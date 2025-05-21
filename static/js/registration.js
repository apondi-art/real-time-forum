function renderRegistrationForm(){
    document.getElementById('app').innerHTML = `
     <div class="register-container">
            <h1>Create Your Account</h1>
            <form id="registerForm">
                <div class="form-group">
                    <label for="nickname">Nickname</label>
                    <input type="text" id="nickname" required placeholder="Choose a display name">
                </div>
                
                <div class="form-row">
                    <div class="form-group" style="flex: 1; margin-right: 15px;">
                        <label for="age">Age</label>
                        <input type="number" id="age" min="13" required placeholder="Your age">
                    </div>
                    <div class="form-group" style="flex: 2;">
                        <label for="gender">Gender</label>
                        <select id="gender" required>
                            <option value="">Select gender</option>
                            <option value="male">Male</option>
                            <option value="female">Female</option>
                            <option value="other">Other</option>
                            <option value="prefer-not-to-say">Prefer not to say</option>
                        </select>
                    </div>
                </div>
                
                <div class="form-row">
                    <div class="form-group" style="flex: 1; margin-right: 15px;">
                        <label for="firstName">First Name</label>
                        <input type="text" id="firstName" required placeholder="Your first name">
                    </div>
                    <div class="form-group" style="flex: 1;">
                        <label for="lastName">Last Name</label>
                        <input type="text" id="lastName" required placeholder="Your last name">
                    </div>
                </div>
                
                <div class="form-group">
                    <label for="email">Email</label>
                    <input type="email" id="email" required placeholder="your@email.com">
                </div>
                
                <div class="form-group">
                    <label for="registerPassword">Password</label>
                    <input type="password" id="registerPassword" required placeholder="Create a password">
                </div>
                
                <div class="form-group">
                    <label for="confirmPassword">Confirm Password</label>
                    <input type="password" id="confirmPassword" required placeholder="Repeat your password">
                </div>
                
                <button type="submit" class="btn">Create Account</button>
                <div id="registerError" class="error-message" style="display: none;"></div>
            </form>
            
            <div class="links">
                <p>Already have an account? <a href="#" id="loginLink">Sign in</a></p>
            </div>
        </div>
    `;  // Handle registration form submission
    document.getElementById('registerForm').addEventListener('submit', function(e) {
        e.preventDefault();
        handleRegistration();
    });

    // Handle login link click
    document.getElementById('loginLink').addEventListener('click', function(e) {
        e.preventDefault();
        renderLoginPage();
    });

}

function handleRegistration() {
    const nickname = document.getElementById('nickname').value;
    const age = document.getElementById('age').value;
    const gender = document.getElementById('gender').value;
    const firstName = document.getElementById('firstName').value;
    const lastName = document.getElementById('lastName').value;
    const email = document.getElementById('email').value;
    const password = document.getElementById('registerPassword').value;
    const confirmPassword = document.getElementById('confirmPassword').value;

    // Basic validation
    if (!nickname || !age || !gender || !firstName || !lastName || !email || !password) {
        showRegisterError('Please fill in all fields');
        return;
    }

    if (password !== confirmPassword) {
        showRegisterError('Passwords do not match');
        return;
    }

    if (age < 13) {
        showRegisterError('You must be at least 13 years old to register');
        return;
    }

    // Registration API call
    fetch('/api/register', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({
            nickname: nickname,
            age: parseInt(age),
            gender: gender,
            firstName: firstName,
            lastName: lastName,
            email: email,
            password: password
        })
    })
    .then(response => {
        if (!response.ok) {
            return response.json().then(data => {
                throw new Error(data.message || 'Registration failed');
            });
        }
        return response.json();
    })
    .then(data => {
        // Registration successful, show login form
        renderLoginPage();
        // Display success message
        setTimeout(() => {
            const errorElement = document.getElementById('errorMessage');
            if (errorElement) {
                errorElement.textContent = 'Registration successful! Please login.';
                errorElement.style.display = 'block';
                errorElement.style.color = 'green';
            }
        }, 100);
    })
    .catch(error => {
        showRegisterError(error.message || 'Registration failed. Please try again.');
    });
}

function showRegisterError(message) {
    const errorElement = document.getElementById('registerError');
    errorElement.textContent = message;
    errorElement.style.display = 'block';
}

// Export the function for other scripts to use
window.renderRegistrationForm = renderRegistrationForm;