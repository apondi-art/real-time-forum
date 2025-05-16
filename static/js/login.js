document.addEventListener('DOMContentLoaded', function(){
    document.getElementById('app').innerHTML = `
    <div class="login-container">
            <h1>Login</h1>
            <form id="loginForm">
                <div class="form-group">
                    <label for="username">Nickname or Email</label>
                    <input type="text" id="username" required>
                </div>
                <div class="form-group">
                    <label for="password">Password</label>
                    <input type="password" id="password" required>
                </div>
                <button type="submit">Login</button>
                <div id="errorMessage" class="error-mesage" style="display: none;"></div>
            </form>
            <div class="links">
                <a href="/register" id="registerLink">Create account</a>
            </div>
        </div>
    `;

    document.getElementById('loginForm').addEventListener('submit', function(e){
        e.preventDefault(); // prevents page reload

        const username = document.getElementById('username').value;
        const password = document.getElementById('password').value;

        if (!username || !password) {
            showError('Please fill in all fields');
            return;
        }

        function showError(message) {
            const errorElement = document.getElementById('errorMessage');
            errorElement.textContent = message;
            errorElement.style.dispaly = 'block';
        }
    })
})