// console.log('login.js loaded'); // Add this line at the beginning of the file

// document.getElementById('login-form').addEventListener('submit', async (e) => {
//     e.preventDefault();

//     const identifier = document.getElementById('identifier').value;
//     const password = document.getElementById('password').value;

//     console.log('Submitting login form'); // Add this line for debugging

//     try {
//         const response = await fetch('/api/login', {
//             method: 'POST',
//             headers: {
//                 'Content-Type': 'application/json',
//             },
//             body: JSON.stringify({ identifier, password }),
//         });

//         if (!response.ok) {
//             throw new Error(`HTTP error: ${response.status}`);
//         }

//         const data = await response.json();
//         const token = data.token;

//         console.log('Received token:', token); // Add this line for debugging

//         // Set the JWT token as a cookie and redirect to the dashboard page
//         document.cookie = `jwt_token=${token};path=/`;
//         window.location.href = '/dashboard';
//     } catch (error) {
//         console.error('Login failed:', error.message); // Update this line for debugging
//         alert('Login failed: ' + error.message);
//     }
// });
