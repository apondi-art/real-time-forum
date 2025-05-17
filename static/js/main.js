
function loadMainApplication() {
    document.getElementById('app').innerHTML = `
        <div class="main-container">
            <header>
                <h1>Forum</h1>
                <button id="logout-btn">Logout</button>
            </header>

            <section id="create-post-section" class="forum-section">
                <h2>Create New Post</h2>
                <form id="new-post-form">
                    <div>
                        <label for="post-title">Title:</label>
                        <input type="text" id="post-title" name="title" required>
                    </div>
                    <div>
                        <label for="post-content">Content:</label>
                        <textarea id="post-content" name="content" rows="5" required></textarea>
                    </div>
                    <button type="submit">Post</button>
                </form>
                <div id="post-creation-message" class="message"></div>
            </section>

            <div class="sections-container">
                <section id="categories-section" class="forum-section">
                    <h2>Categories</h2>
                    <div id="categories-list"></div>
                </section>

                <section id="posts-section" class="forum-section">
                    <h2>Posts</h2>
                    <div id="posts-feed"></div>
                </section>

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
    document.getElementById('new-post-form').addEventListener('submit', handleCreatePost);
}

// Handle post creation
function handleCreatePost(event) {
    event.preventDefault(); // Prevent default form submission

    const title = document.getElementById('post-title').value;
    const content = document.getElementById('post-content').value;
    const messageDiv = document.getElementById('post-creation-message');
    const token = localStorage.getItem('forum_token'); // Use correct token key

    if (!token) {
        messageDiv.textContent = 'You must be logged in to create a post.';
        messageDiv.className = 'message error';
        return;
    }

    fetch('/api/posts', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${token}`, // Include the JWT token
        },
        body: JSON.stringify({ title, content }),
    })
    .then(response => {
        if (!response.ok) {
            return response.json().then(data => {
                throw new Error(data.message || `Error: ${response.status}`);
            });
        }
        return response.json();
    })
    .then(data => {
        if (data.success) {
            messageDiv.textContent = 'Post created successfully!';
            messageDiv.className = 'message success';
            
            // Reset the form
            document.getElementById('new-post-form').reset();
            
          
            setTimeout(() => {
                loadPosts(); // Reload the posts
            }, 500);
        } else {
            messageDiv.textContent = `Failed to create post: ${data.message}`;
            messageDiv.className = 'message error';
        }
    })
    .catch(error => {
        console.error('Error creating post:', error);
        messageDiv.textContent = 'An error occurred while creating the post.';
        messageDiv.className = 'message error';
    });
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
    const token = localStorage.getItem('forum_token'); // Use consistent token name
    
    fetch('/api/posts', {
        headers: {
            'Authorization': `Bearer ${token}` // Include authentication token
        }
    })
    .then(response => {
        if (!response.ok) {
            throw new Error(`Failed to load posts: ${response.status}`);
        }
        return response.json();
    })
    .then(posts => {
        const container = document.getElementById('posts-feed');
        
        if (!posts || posts.length === 0) {
            container.innerHTML = '<p>No posts found. Be the first to create one!</p>';
            return;
        }
        
        container.innerHTML = posts.map(post => `
            <div class="post" data-id="${post.id || post.ID}">
                <h3>${post.title}</h3>
                <p>${post.content}</p>
                <div class="post-meta">
                    <span class="post-author">Posted by: ${post.author || 'Anonymous'}</span>
                    <span class="post-date">${formatDate(post.created_at || post.createdAt || new Date())}</span>
                </div>
            </div>
        `).join('');
    })
    .catch(error => {
        console.error('Error loading posts:', error);
        document.getElementById('posts-feed').innerHTML = 
            `<p class="error">Failed to load posts. Please try again later.</p>`;
    });
}

// Helper function to format dates nicely
function formatDate(dateString) {
    const date = new Date(dateString);
    if (isNaN(date.getTime())) {
        return 'Unknown date';
    }
    return date.toLocaleString();
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
    const token = localStorage.getItem('forum_token'); 
    
    if (!token) {
        console.log('No token found, already logged out');
        window.location.href = '/';
        return;
    }
    
    fetch('/api/logout', { 
        method: 'POST',
        headers: {
            'Authorization': `Bearer ${token}` // Include token in header
        }
    })
    .then(response => {
        // Handle non-JSON responses gracefully
        if (response.ok) {
            try {
                return response.json();
            } catch (e) {
                return { success: true }; // Assume success if server returned 200 OK
            }
        } else {
            console.error('Logout request failed with status:', response.status);
            return { success: false, message: 'Server returned error status: ' + response.status };
        }
    })
    .then(data => {
        // Clear all user data regardless of server response
        localStorage.removeItem('forum_token');
        localStorage.removeItem('user_id');
        localStorage.removeItem('username');
        
        console.log('Logout complete, redirecting to login page');
        
        // Force page reload to ensure clean application state
        window.location.reload();
    })
    .catch(error => {
        console.error('Error during logout:', error);
        alert('Logout failed. Please try again or refresh the page.');
    });
}

window.loadMainApplication = loadMainApplication;