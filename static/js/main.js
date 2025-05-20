function loadMainApplication() {
    document.getElementById('app').innerHTML = `
        <div class="main-container">
            <header>
                <h1>Forum</h1>
                <button id="logout-btn">Logout</button>
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
            <!-- Post Details Modal with Comments Section -->
            <div id="post-detail-modal" class="modal" style="display:none;">
                <div class="modal-content">
                    <span class="close-modal">&times;</span>
                    <div id="post-detail-content"></div>
                    
                    <div class="comments-section">
                        <h3>Comments</h3>
                        <div id="comments-container"></div>
                        
                        <form id="add-comment-form">
                            <textarea id="comment-content" placeholder="Write a comment..." required></textarea>
                            <button type="submit">Add Comment</button>
                        </form>
                    </div>
                </div>
            </div>
        </div>
    `;

    // Load initial data
    loadCategories();
    loadPosts();
    loadOnlineUsers();

    // Add event listeners
    document.getElementById('logout-btn').addEventListener('click', handleLogout);
    document.getElementById('create-post-button').addEventListener('click', showCreatePostForm);
    document.getElementById('new-post-form').addEventListener('submit', handleCreatePost);

      // Close modal when clicking on X
      document.querySelector('.close-modal').addEventListener('click', closePostDetailModal);
    
      // Add comment form submission
      document.getElementById('add-comment-form').addEventListener('submit', handleAddComment);
     // Add periodic status updates
     updateOnlineStatus(true);
     setInterval(() => updateOnlineStatus(true), 30000); // Update every 30 seconds
     
     // Add beforeunload event to mark user as offline when leaving
     window.addEventListener('beforeunload', () => {
         updateOnlineStatus(false);
     });
}

function showCreatePostForm() {
    document.getElementById('create-post-section').style.display = 'block';
    document.getElementById('create-post-button-section').style.display = 'none';
}

function closePostDetailModal() {
    document.getElementById('post-detail-modal').style.display = 'none';
}
// Handle post creation
function handleCreatePost(event) {
    event.preventDefault(); // Prevent default form submission

    const title = document.getElementById('post-title').value;
    const content = document.getElementById('post-content').value;
     const category = document.getElementById('post-category').value;
    const messageDiv = document.getElementById('post-creation-message');
    const token = localStorage.getItem('forum_token'); // Use correct token key

    if (!token) {
        messageDiv.textContent = 'You must be logged in to create a post.';
        messageDiv.className = 'message error';
        return;
    }

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
        body: JSON.stringify({ title, content ,category}),
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
                loadCategories(); // Reload categories to update countsm
                document.getElementById('create-post-section').style.display = 'none';
                document.getElementById('create-post-button-section').style.display = 'block';
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
    const token = localStorage.getItem('forum_token'); // Use consistent token name
    
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
                </div>
            </div>
        `).join('');

          // Add click event to posts to open detail modal
        document.querySelectorAll('.post').forEach(post => {
            post.addEventListener('click', () => {
                const postId = post.dataset.id;
                openPostDetailModal(postId);
            });
        });
    })
    .catch(error => {
        console.error('Error loading posts:', error);
        document.getElementById('posts-feed').innerHTML = 
            `<p class="error">Failed to load posts: ${error.message}</p>`;
    });
}


// Function to open post detail modal with comments
function openPostDetailModal(postId) {
    // Get post details
    const token = localStorage.getItem('forum_token');

     // First verify the postId exists and is valid
    if (!postId) {
        console.error('No post ID provided');
        return;
    }

    console.log(`Attempting to load post with ID: ${postId}`); // Debug log
    
    fetch(`/api/posts/${postId}`, {
        headers: {
            'Authorization': `Bearer ${token}`
        }
    })
    
    .then(response => {
        if (!response.ok) throw new Error(`Failed to load post: ${response.status}`);
        return response.json();
    })
    .then(post => {
        // Display post content in modal
        document.getElementById('post-detail-content').innerHTML = `
            <div class="post-full">
                <h2>${post.title}</h2>
                <p>${post.content}</p>
                <div class="post-meta">
                    <span class="post-category">Category: ${post.category || 'Uncategorized'}</span>
                    <span class="post-author">Posted by: ${post.author || 'Anonymous'}</span>
                    <span class="post-date">${formatDate(post.created_at || post.createdAt || new Date())}</span>
                </div>
            </div>
        `;
        
        // Store current post ID in the form for comment submission
        document.getElementById('add-comment-form').dataset.postId = postId;
        
        // Load comments for this post
        loadComments(postId);
        
        // Show the modal
        document.getElementById('post-detail-modal').style.display = 'block';
    })
    .catch(error => {
        console.error('Error loading post details:', error);
        alert('Failed to load post details. Please try again.');
    });
}

// Function to load comments for a specific post
function loadComments(postId) {
    const token = localStorage.getItem('forum_token');
    
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
        const container = document.getElementById('comments-container');
        
        if (!comments || comments.length === 0) {
            container.innerHTML = '<p>No comments yet. Be the first to comment!</p>';
            return;
        }
        
        container.innerHTML = comments.map(comment => `
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
        console.error('Error loading comments:', error);
        document.getElementById('comments-container').innerHTML = 
            `<p class="error">Failed to load comments: ${error.message}</p>`;
    });
}

// Function to handle adding a new comment
function handleAddComment(event) {
    event.preventDefault();
    
    const token = localStorage.getItem('forum_token');
    const postId = event.target.dataset.postId;
    const content = document.getElementById('comment-content').value.trim();
    
    if (!token) {
        alert('You must be logged in to comment.');
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
            document.getElementById('comment-content').value = '';
            
            // Reload comments to show the new one
            loadComments(postId);
            
            // Update post's comment count in the main view
            const commentButton = document.querySelector(`.view-comments-btn[data-id="${postId}"] .comment-count`);
            if (commentButton) {
                const currentCount = parseInt(commentButton.textContent) || 0;
                commentButton.textContent = currentCount + 1;
            }
        } else {
            alert(`Failed to add comment: ${data.message}`);
        }
    })
    .catch(error => {
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
            
            container.innerHTML = users.map(user => `
                <div class="user" data-id="${user.id}">
                    ${user.username} 
                    <span class="status ${user.online ? 'online' : 'offline'}"></span>
                </div>
            `).join('');
        })
        .catch(error => {
            console.error('Error loading online users:', error);
            document.getElementById('users-list').innerHTML = 
                `<div class="error">Failed to load online users</div>`;
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
    .catch(error => {// Function to handle adding a new comment
        function handleAddComment(event) {
            event.preventDefault();
            
            const token = localStorage.getItem('forum_token');
            const postId = event.target.dataset.postId;
            const content = document.getElementById('comment-content').value.trim();
            
            if (!token) {
                alert('You must be logged in to comment.');
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
                    document.getElementById('comment-content').value = '';
                    
                    // Reload comments to show the new one
                    loadComments(postId);
                    
                    // Update post's comment count in the main view
                    const commentButton = document.querySelector(`.view-comments-btn[data-id="${postId}"] .comment-count`);
                    if (commentButton) {
                        const currentCount = parseInt(commentButton.textContent) || 0;
                        commentButton.textContent = currentCount + 1;
                    }
                } else {
                    alert(`Failed to add comment: ${data.message}`);
                }
            })
            .catch(error => {
                console.error('Error adding comment:', error);
                alert('An error occurred while adding your comment. Please try again.');
            });
        }
        console.error('Error during logout:', error);
        alert('Logout failed. Please try again or refresh the page.');
    });
}

// Update the online status
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

// Enhanced loadOnlineUsers function
function loadOnlineUsers() {
    fetch('/api/online-users')
        .then(response => {
            if (!response.ok) {
                throw new Error(`Failed to load users: ${response.status}`);
            }
            return response.json();
        })
        .then(users => {
            const container = document.getElementById('users-list');
            
            if (!users || !Array.isArray(users) || users.length === 0) {
                container.innerHTML = '<div>No users found</div>';
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
        })
        .catch(error => {
            console.error('Error loading users:', error);
            document.getElementById('users-list').innerHTML = 
                `<div class="error">Failed to load users</div>`;
        });
}

function createUserElement(user) {
    const lastSeen = user.lastSeen ? formatLastSeen(user.lastSeen) : '';
    return `
        <div class="user ${user.online ? 'online' : 'offline'}" data-id="${user.id}">
            <div class="user-info">
                <div class="user-name">${user.nickname}</div>
                <div class="user-status">
                    ${user.online ? 'Online' : `Last seen ${lastSeen}`}
                </div>
            </div>
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

window.loadMainApplication = loadMainApplication;
