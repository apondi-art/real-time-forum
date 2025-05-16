function loadMainApplication() {
    document.getElementById('app').innerHTML = `
        <div class="main-container">
            <!-- Header (e.g., logout button) -->
            <header>
                <h1>Forum</h1>
                <button id="logout-btn">Logout</button>
            </header>

            <!-- Three Subsections -->
            <div class="sections-container">
                <!-- 1. Categories Section -->
                <section id="categories-section" class="forum-section">
                    <h2>Categories</h2>
                    <div id="categories-list"></div>
                </section>

                <!-- 2. Posts Section -->
                <section id="posts-section" class="forum-section">
                    <h2>Posts</h2>
                    <div id="posts-feed"></div>
                </section>

                <!-- 3. Users Section (for private messages) -->
                <section id="users-section" class="forum-section">
                    <h2>Online Users</h2>
                    <div id="users-list"></div>
                </section>
            </div>
        </div>
    `;

    // Load initial data
    loadCategories();
    loadPosts();
    loadOnlineUsers();

    // Add event listeners
    document.getElementById('logout-btn').addEventListener('click', handleLogout);
}




// Load categories from backend
function loadCategories() {
    fetch('/api/categories')
        .then(response => response.json())
        .then(categories => {
            const container = document.getElementById('categories-list');
            container.innerHTML = categories.map(cat => `
                <div class="category" data-id="${cat.id}">
                    ${cat.name} (${cat.postCount})
                </div>
            `).join('');
        });
}

// Load posts
function loadPosts() {
    fetch('/api/posts')
        .then(response => response.json())
        .then(posts => {
            const container = document.getElementById('posts-feed');
            container.innerHTML = posts.map(post => `
                <div class="post" data-id="${post.id}">
                    <h3>${post.title}</h3>
                    <p>${post.content}</p>
                </div>
            `).join('');
        });
}

// Load online users
function loadOnlineUsers() {
    fetch('/api/online-users')
        .then(response => response.json())
        .then(users => {
            const container = document.getElementById('users-list');
            container.innerHTML = users.map(user => `
                <div class="user" data-id="${user.id}">
                    ${user.username} 
                    <span class="status ${user.online ? 'online' : 'offline'}"></span>
                </div>
            `).join('');
        });
}

// Logout handler
function handleLogout() {
    fetch('/logout', { method: 'POST' })
        .then(() => window.location.reload());
}

window.loadMainApplication = loadMainApplication;