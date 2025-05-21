// Global variable for token check interval
let tokenCheckInterval;

function loadMainApplication() {
    document.getElementById('app').innerHTML = `
        <div class="main-container">
            <header>
                <h1>Forum</h1>
                <div class="header-icons">
                    <button id="notification-icon" class="icon-button">
                        <i class="fas fa-bell"></i> <span class="notification-badge" style="display: none;"></span>
                    </button>
                    <button id="logout-btn">Logout</button>
                </div>
            </header>
            
            <section id="create-post-button-section" class="forum-section">
                <button id="create-post-button">Create Post</button>
            </section>

            <section id="create-post-section" class="forum-section" style="display: none;">
                <h2>Create New Post</h2>
                <form id="new-post-form">
                    <div>
                        <label for="post-title">Title:</label>
                        <input type="text" id="post-title" name="title" required>
                    </div>
                    <div>
                        <label for="post-category">Category:</label>
                        <select id="post-category" name="category" required>
                            <option value="">-- Select a Category --</option>
                            <option value="Sports">Sports</option>
                            <option value="Lifestyle">Lifestyle</option>
                            <option value="Education">Education</option>
                            <option value="Finance">Finance</option>
                            <option value="Music">Music</option>
                            <option value="Culture">Culture</option>
                            <option value="Technology">Technology</option>
                            <option value="Health">Health</option>
                            <option value="Travel">Travel</option>
                            <option value="Food">Food</option>
                        </select>
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

            <div id="chat-window" class="chat-window" style="display: none;">
                <div class="chat-header">
                    <h3 id="chat-header-title">Chat</h3>
                    <button class="close-chat-btn">&times;</button>
                </div>
                <div class="chat-messages" id="chat-messages">
                    </div>
                <form id="chat-form" class="chat-input-form">
                    <input type="text" id="chat-input" placeholder="Type a message..." autocomplete="off">
                    <button type="submit">Send</button>
                </form>
            </div>
        </div>
    `;

    // Load initial data
    loadCategories();
    loadPosts();
    loadOnlineUsers();

    // Add event listeners
    document.getElementById('logout-btn').addEventListener('click', handleLogout);
    document.getElementById('create-post-button').addEventListener('click', toggleCreatePostForm);
    document.getElementById('new-post-form').addEventListener('submit', handleCreatePost);
    
    // Add periodic status updates
    updateOnlineStatus(true);
    setInterval(() => updateOnlineStatus(true), 30000); // Update every 30 seconds
     
    // Add beforeunload event to mark user as offline when leaving
    window.addEventListener('beforeunload', () => {
        updateOnlineStatus(false);
        clearInterval(tokenCheckInterval);
    });
    
    // Start token validation
    startTokenValidation();

    // Setup global error handling
    setupGlobalErrorHandling();

    // Event listener for notification icon
    document.getElementById('notification-icon').addEventListener('click', () => {
        // This will be expanded to open a messages/notifications panel
        alert('Notifications clicked! (Feature to be implemented)');
        // For now, let's assume clicking it clears the badge
        updateNotificationBadge(0); 
    });

    // Event listeners for chat window
    document.querySelector('.close-chat-btn').addEventListener('click', () => {
        document.getElementById('chat-window').style.display = 'none';
    });

    document.getElementById('chat-form').addEventListener('submit', handleSendMessage);

    // Simulate a new notification
    setTimeout(() => updateNotificationBadge(3), 5000); 
}

// Token validation functions 
function startTokenValidation() {
    setTimeout(() => {
        checkTokenValidity();
        // set regular interval checks
        tokenCheckInterval = setInterval(checkTokenValidity, 5 * 60 * 1000);
    }, 2000); // 2-second delay before first check
}

function checkTokenValidity() {
    const token = localStorage.getItem('forum_token');
    if (!token) return;

    fetch('/api/validate-token', {
        headers: {
            'Authorization': `Bearer ${token}`
        }
    })
    .then(response => {
        // Only invalidate on specific unauthorized responses (401)
        // Other error codes (like 500, 502, etc.) should be treated as temporary server issues
        if (response.status === 401) {
            clearInterval(tokenCheckInterval);
            handleInvalidToken();
        }
        // For server errors, log but don't invalidate token
        else if (!response.ok) {
            console.warn('Token validation server error:', response.status);
            // Don't invalidate on server errors
        }
    })
    .catch(error => {
        // For network errors, just log the error but don't invalidate the token
        // This prevents users from being logged out due to temporary connection issues
        console.error('Token validation network error:', error);
        // Don't invalidate on network errors
    });
}

function handleInvalidToken() {
    // Add some debug logging
    console.log('Token invalidated. Logging out user.');
    
    // Add a check to prevent multiple logout attempts
    if (!localStorage.getItem('forum_token')) {
        console.log('Already logged out. Skipping additional logout.');
        return;
    }
    
    // Clear all user data
    localStorage.removeItem('forum_token');
    localStorage.removeItem('user_id');
    localStorage.removeItem('username');
    
    // Show a message to the user
    alert('Your session has expired. Please log in again.');
    
    // Redirect to login page
    window.location.href = '/';
}

function setupGlobalErrorHandling() {
    // Intercept fetch calls to check for 401 errors
    const originalFetch = window.fetch;
    
    window.fetch = async function(...args) {
        try {
            const response = await originalFetch(...args);
            
            // Only treat actual 401 responses as token invalidation events
            if (response.status === 401) {
                // Check if this is a token validation endpoint
                const url = args[0] instanceof Request ? args[0].url : args[0];
                
                // Unauthorized - token is invalid
                clearInterval(tokenCheckInterval);
                handleInvalidToken();
                return Promise.reject(new Error('Unauthorized'));
            }
            
            return response;
        } catch (error) {
            // Only handle actual unauthorized errors, not network failures
            if (error.message === 'Unauthorized') {
                clearInterval(tokenCheckInterval);
                handleInvalidToken();
            }
            return Promise.reject(error);
        }
    };
}

function toggleCreatePostForm() {
    const formSection = document.getElementById('create-post-section');
    const createButton = document.getElementById('create-post-button');
    
    if (formSection.style.display === 'none') {
        // Show the form
        formSection.style.display = 'block';
        createButton.textContent = 'Cancel';
        createButton.setAttribute('data-mode', 'cancel');
    } else {
        // Hide the form
        formSection.style.display = 'none';
        createButton.textContent = 'Create Post';
        createButton.removeAttribute('data-mode');
        
        // Clear any validation messages
        const messageDiv = document.getElementById('post-creation-message');
        if (messageDiv) {
            messageDiv.textContent = '';
            messageDiv.className = 'message';
        }
        
        // Optionally reset the form
        document.getElementById('new-post-form').reset();
    }
}

// Handle post creation
function handleCreatePost(event) {
    event.preventDefault(); // Prevent default form submission

    const token = localStorage.getItem('forum_token');
    if (!token) {
        handleInvalidToken();
        return;
    }

    const title = document.getElementById('post-title').value;
    const content = document.getElementById('post-content').value;
    const category = document.getElementById('post-category').value;
    const messageDiv = document.getElementById('post-creation-message');

    if (!category) {
        messageDiv.textContent = 'Please select a category for your post.';
        messageDiv.className = 'message error';
        return;
    }

    fetch('/api/posts', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${token}`, // Include the JWT token
        },
        body: JSON.stringify({ title, content, category }),
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
                loadCategories(); // Reload categories to update counts
                document.getElementById('create-post-section').style.display = 'none';
                document.getElementById('create-post-button-section').style.display = 'block';
            }, 500);
        } else {
            messageDiv.textContent = `Failed to create post: ${data.message}`;
            messageDiv.className = 'message error';
        }
    })
    .catch(error => {
        if (error.message === 'Unauthorized') return;
        console.error('Error creating post:', error);
        messageDiv.textContent = 'An error occurred while creating the post.';
        messageDiv.className = 'message error';
    });
}

// Load categories from backend
function loadCategories() {
    fetch('/api/categories')
        .then(response => {
            if (!response.ok) {
                throw new Error(`Failed to load categories: ${response.status}`);
            }
            return response.json();
        })
        .then(categories => {
            // Check if categories is null or not an array
            if (!categories || !Array.isArray(categories)) {
                throw new Error('Invalid categories data received');
            }
            
            const container = document.getElementById('categories-list');
            if (categories.length === 0) {
                container.innerHTML = '<div class="category-empty">No categories available</div>';
            } else {
                container.innerHTML = categories.map(cat => `
                    <div class="category" data-id="${cat.id}">
                        ${cat.name}
                    </div>
                `).join('');
                
                // Add click event to filter posts by category
                document.querySelectorAll('.category').forEach(categoryEl => {
                    categoryEl.addEventListener('click', () => {
                        const categoryId = categoryEl.dataset.id;
                        loadPosts(categoryId); // Add categoryId parameter to loadPosts
                    });
                });
            }
        })
        .catch(error => {
            console.error('Error loading categories:', error);
            document.getElementById('categories-list').innerHTML = 
                `<div class="error">Failed to load categories: ${error.message}</div>`;
        });
}

function loadPosts(categoryId = null) {
    const token = localStorage.getItem('forum_token');
    if (!token) {
        handleInvalidToken();
        return;
    }
    
    // Build URL with query parameter if categoryId is provided
    let url = '/api/posts';
    if (categoryId) {
        url += `?category=${categoryId}`;
    }
    
    fetch(url, {
        headers: {
            'Authorization': `Bearer ${token}` // Include authentication token
        }
    })
    .then(response => {
        if (!response.ok) {
            return response.text().then(text => {
                try {
                    const data = JSON.parse(text);
                    throw new Error(data.message || `Error: ${response.status}`);
                } catch (e) {
                    throw new Error(`Failed to load posts: ${response.status}`);
                }
            });
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
                    <span class="post-category">Category: ${post.category || 'Uncategorized'}</span></br>
                    <span class="post-author">Posted by: ${post.author || 'Anonymous'}</span></br>
                    <span class="post-date">${formatDate(post.created_at || post.createdAt || new Date())}</span>
                    <div class="post-actions">
                        <button class="comment-btn" data-id="${post.id || post.ID}">
                            <i class="comment-icon">💬</i> Comments
                        </button>
                    </div>
                </div>
                <div class="comments-section" id="comments-section-${post.id || post.ID}" style="display: none;">
                    <div class="comments-container" id="comments-container-${post.id || post.ID}">
                        <p>Loading comments...</p>
                    </div>
                    <form class="add-comment-form" data-post-id="${post.id || post.ID}">
                        <textarea class="comment-content" placeholder="Write a comment..." required></textarea>
                        <button type="submit">Add Comment</button>
                    </form>
                </div>
            </div>
        `).join('');

        // Add click event to comment buttons
        document.querySelectorAll('.comment-btn').forEach(btn => {
            btn.addEventListener('click', (event) => {
                event.stopPropagation(); // Prevent post click event
                const postId = btn.dataset.id;
                toggleComments(postId);
            });
        });
        
        // Add submit event to comment forms
        document.querySelectorAll('.add-comment-form').forEach(form => {
            form.addEventListener('submit', (event) => {
                event.preventDefault();
                event.stopPropagation(); // Prevent post click event
                const postId = form.dataset.postId;
                const content = form.querySelector('.comment-content').value.trim();
                addComment(postId, content, form);
            });
        });
        
        // Add click event to posts for viewing details
        document.querySelectorAll('.post').forEach(post => {
            post.addEventListener('click', (event) => {
                // Don't trigger if clicking on comment button or form
                if (!event.target.closest('.comments-section') && 
                    !event.target.closest('.comment-btn')) {
                    const postId = post.dataset.id;
                    viewPostDetails(postId);
                }
            });
        });
    })
    .catch(error => {
        if (error.message.includes('Unauthorized')) return;
        console.error('Error loading posts:', error);
        document.getElementById('posts-feed').innerHTML = 
            `<p class="error">Failed to load posts: ${error.message}</p>`;
    });
}

// Function to toggle comments visibility and load them if needed
function toggleComments(postId) {
    const commentsSection = document.getElementById(`comments-section-${postId}`);
    
    if (commentsSection.style.display === 'none') {
        commentsSection.style.display = 'block';
        loadComments(postId);
    } else {
        commentsSection.style.display = 'none';
    }
}

// Function to view full post details
function viewPostDetails(postId) {
    // You can either keep the original modal or implement an inline expansion
    console.log(`View details for post: ${postId}`);
    // For now, just toggle comments to show more info
    toggleComments(postId);
}

// Function to load comments for a specific post
function loadComments(postId) {
    const token = localStorage.getItem('forum_token');
    if (!token) {
        handleInvalidToken();
        return;
    }
    
    const commentsContainer = document.getElementById(`comments-container-${postId}`);
    
    fetch(`/api/posts/${postId}/comments`, {
        headers: {
            'Authorization': `Bearer ${token}`
        }
    })
    .then(response => {
        if (!response.ok) throw new Error(`Failed to load comments: ${response.status}`);
        return response.json();
    })
    .then(comments => {
        if (!comments || comments.length === 0) {
            commentsContainer.innerHTML = '<p>No comments yet. Be the first to comment!</p>';
            return;
        }
        
        commentsContainer.innerHTML = comments.map(comment => `
            <div class="comment">
                <div class="comment-content">${comment.content}</div>
                <div class="comment-meta">
                    <span class="comment-author">By: ${comment.author || 'Anonymous'}</span>
                    <span class="comment-date">${formatDate(comment.created_at || comment.createdAt)}</span>
                </div>
            </div>
        `).join('');
    })
    .catch(error => {
        if (error.message.includes('Unauthorized')) return;
        console.error('Error loading comments:', error);
        commentsContainer.innerHTML = `<p class="error">Failed to load comments: ${error.message}</p>`;
    });
}

// Function to add a new comment
function addComment(postId, content, form) {
    const token = localStorage.getItem('forum_token');
    
    if (!token) {
        handleInvalidToken();
        return;
    }
    
    if (!content) {
        alert('Comment cannot be empty.');
        return;
    }
    
    fetch(`/api/posts/${postId}/comments`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${token}`
        },
        body: JSON.stringify({ content })
    })
    .then(response => {
        if (!response.ok) throw new Error(`Failed to add comment: ${response.status}`);
        return response.json();
    })
    .then(data => {
        if (data.success) {
            // Clear the comment form
            form.querySelector('.comment-content').value = '';
            
            // Reload comments to show the new one
            loadComments(postId);
        } else {
            alert(`Failed to add comment: ${data.message}`);
        }
    })
    .catch(error => {
        if (error.message.includes('Unauthorized')) return;
        console.error('Error adding comment:', error);
        alert('An error occurred while adding your comment. Please try again.');
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
        .then(response => {
            if (!response.ok) {
                throw new Error(`Failed to load online users: ${response.status}`);
            }
            return response.json();
        })
        .then(users => {
            const container = document.getElementById('users-list');
            
            if (!users || !Array.isArray(users) || users.length === 0) {
                container.innerHTML = '<div>No users online</div>';
                return;
            }
            
            // Separate online and offline users
            const onlineUsers = users.filter(u => u.online);
            const offlineUsers = users.filter(u => !u.online);
            
            let html = '';
            
            if (onlineUsers.length > 0) {
                html += `
                    <div class="user-group">
                        <div class="user-group-title">Online - ${onlineUsers.length}</div>
                        ${onlineUsers.map(user => createUserElement(user)).join('')}
                    </div>
                `;
            }
            
            if (offlineUsers.length > 0) {
                html += `
                    <div class="user-group">
                        <div class="user-group-title">Offline - ${offlineUsers.length}</div>
                        ${offlineUsers.map(user => createUserElement(user)).join('')}
                    </div>
                `;
            }
            
            container.innerHTML = html;
            addChatIconListeners(); // Add listeners after users are loaded
        })
        .catch(error => {
            console.error('Error loading users:', error);
            document.getElementById('users-list').innerHTML = 
                `<div class="error">Failed to load users</div>`;
        });
}

function createUserElement(user) {
    const lastSeen = user.lastSeen ? formatLastSeen(user.lastSeen) : '';
    const userId = localStorage.getItem('user_id'); // Get the current user's ID
    const isCurrentUser = (userId && parseInt(userId) === user.id); // Check if it's the current user
    
    return `
        <div class="user ${user.online ? 'online' : 'offline'}" data-id="${user.id}" data-username="${user.nickname}">
            <div class="user-info">
                <div class="user-name">${user.nickname}</div>
                <div class="user-status">
                    ${user.online ? 'Online' : `Last seen ${lastSeen}`}
                </div>
            </div>
            ${!isCurrentUser ? `<button class="chat-user-icon" data-id="${user.id}" data-username="${user.nickname}">
                <i class="fas fa-comment-dots"></i> </button>` : ''}
        </div>
    `;
}

function formatLastSeen(timestamp) {
    const now = new Date();
    const date = new Date(timestamp);
    const diffHours = Math.floor((now - date) / (1000 * 60 * 60));
    
    if (diffHours < 1) {
        const diffMinutes = Math.floor((now - date) / (1000 * 60));
        return `${diffMinutes} minute${diffMinutes !== 1 ? 's' : ''} ago`;
    } else if (diffHours < 24) {
        return `${diffHours} hour${diffHours !== 1 ? 's' : ''} ago`;
    } else {
        const diffDays = Math.floor(diffHours / 24);
        return `${diffDays} day${diffDays !== 1 ? 's' : ''} ago`;
    }
}

// Update online status
function updateOnlineStatus(online) {
    const token = localStorage.getItem('forum_token');
    if (!token) return;

    fetch('/api/online-status', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({ online }),
    }).catch(err => console.error('Error updating status:', err));
}

// Logout handler
function handleLogout() {
    clearInterval(tokenCheckInterval);
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

// --- Chat Functionality (New) ---

let currentChatRecipientId = null;
let currentChatRecipientUsername = null;

function addChatIconListeners() {
    document.querySelectorAll('.chat-user-icon').forEach(button => {
        button.addEventListener('click', (event) => {
            event.stopPropagation(); // Prevent clicking on the user div from triggering other actions
            const userId = event.currentTarget.dataset.id;
            const username = event.currentTarget.dataset.username;
            openChatWindow(userId, username);
        });
    });
}

function openChatWindow(recipientId, recipientUsername) {
    const chatWindow = document.getElementById('chat-window');
    const chatHeaderTitle = document.getElementById('chat-header-title');
    const chatMessagesContainer = document.getElementById('chat-messages');

    currentChatRecipientId = recipientId;
    currentChatRecipientUsername = recipientUsername;
    chatHeaderTitle.textContent = `Chat with ${recipientUsername}`;
    chatMessagesContainer.innerHTML = '<p>Loading messages...</p>'; // Clear and show loading

    chatWindow.style.display = 'block';
    loadChatMessages(recipientId);
}

function loadChatMessages(recipientId) {
    const token = localStorage.getItem('forum_token');
    if (!token) {
        handleInvalidToken();
        return;
    }

    fetch(`/api/chat/messages?recipientId=${recipientId}`, {
        headers: {
            'Authorization': `Bearer ${token}`
        }
    })
    .then(response => {
        if (!response.ok) throw new Error(`Failed to load chat messages: ${response.status}`);
        return response.json();
    })
    .then(messages => {
        const chatMessagesContainer = document.getElementById('chat-messages');
        chatMessagesContainer.innerHTML = ''; // Clear loading message

        if (messages.length === 0) {
            chatMessagesContainer.innerHTML = '<p>No messages yet. Start the conversation!</p>';
            return;
        }

        const currentUserId = parseInt(localStorage.getItem('user_id'));

        messages.forEach(msg => {
            const messageElement = document.createElement('div');
            messageElement.classList.add('chat-message');
            messageElement.classList.add(msg.sender_id === currentUserId ? 'sent' : 'received');
            
            messageElement.innerHTML = `
                <div class="message-content">${msg.content}</div>
                <div class="message-meta">
                    <span class="message-sender">${msg.sender_username || 'Unknown'}</span>
                    <span class="message-time">${formatChatTime(msg.created_at)}</span>
                </div>
            `;
            chatMessagesContainer.appendChild(messageElement);
        });
        chatMessagesContainer.scrollTop = chatMessagesContainer.scrollHeight; // Scroll to bottom
    })
    .catch(error => {
        console.error('Error loading chat messages:', error);
        document.getElementById('chat-messages').innerHTML = `<p class="error">Failed to load messages: ${error.message}</p>`;
    });
}

function handleSendMessage(event) {
    event.preventDefault();
    const chatInput = document.getElementById('chat-input');
    const content = chatInput.value.trim();

    if (!content || !currentChatRecipientId) {
        return;
    }

    const token = localStorage.getItem('forum_token');
    if (!token) {
        handleInvalidToken();
        return;
    }

    fetch('/api/chat/send', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${token}`
        },
        body: JSON.stringify({ recipientId: currentChatRecipientId, content })
    })
    .then(response => {
        if (!response.ok) throw new Error(`Failed to send message: ${response.status}`);
        return response.json();
    })
    .then(data => {
        if (data.success) {
            chatInput.value = ''; // Clear input
            loadChatMessages(currentChatRecipientId); // Reload messages to show the new one
        } else {
            alert(`Failed to send message: ${data.message}`);
        }
    })
    .catch(error => {
        console.error('Error sending message:', error);
        alert('An error occurred while sending your message.');
    });
}

function updateNotificationBadge(count) {
    const badge = document.querySelector('.notification-badge');
    if (count > 0) {
        badge.textContent = count;
        badge.style.display = 'block';
    } else {
        badge.textContent = '';
        badge.style.display = 'none';
    }
}

function formatChatTime(timestamp) {
    const date = new Date(timestamp);
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
}

window.loadMainApplication = loadMainApplication;