function getToken() { return localStorage.getItem('jwt_token') || ''; }
function getUser() {
    try { return JSON.parse(localStorage.getItem('jwt_user') || 'null'); }
    catch { return null; }
}

async function apiFetch(url, opts = {}) {
    opts.headers = opts.headers || {};
    opts.headers['Content-Type'] = 'application/json';
    const token = getToken();
    if (token) opts.headers['Authorization'] = 'Bearer ' + token;
    const res = await fetch(url, opts);
    return res.json();
}

function renderAuthBar() {
    const user = getUser();
    const bar = document.createElement('div');
    bar.style.cssText = 'background:#0f172a;color:#fff;padding:10px 24px;display:flex;justify-content:space-between;align-items:center;font-size:13px;flex-wrap:wrap;gap:8px;';
    if (user && getToken()) {
        bar.innerHTML = `
            <div>🔑 Вы вошли как: <b>${user.username}</b> (${user.email}) — Роль: <span class="badge badge-ok">${user.role}</span></div>
            <button class="btn-sm btn-red" onclick="logoutUser()">Выйти</button>
        `;
    } else {
        bar.innerHTML = `
            <div>⚠️ Не авторизован. Войдите как Admin: <b>admin@library.local</b> / <b>admin123</b></div>
            <div style="display:flex;gap:6px;align-items:center;">
                <input id="loginEmail" placeholder="Email" value="admin@library.local" style="margin:0;padding:5px 8px;width:170px;">
                <input id="loginPass" type="password" placeholder="Пароль" value="admin123" style="margin:0;padding:5px 8px;width:120px;">
                <button class="btn-sm btn-green" onclick="loginUser()">Войти</button>
                <button class="btn-sm" onclick="registerUser()">Регистрация (USER)</button>
            </div>
        `;
    }
    document.body.insertBefore(bar, document.body.firstChild);
}

async function loginUser() {
    const email = document.getElementById('loginEmail').value;
    const password = document.getElementById('loginPass').value;
    const res = await apiFetch('/api/auth/login', { method: 'POST', body: JSON.stringify({ email, password }) });
    if (res.isSuccess) {
        localStorage.setItem('jwt_token', res.result.token);
        localStorage.setItem('jwt_user', JSON.stringify(res.result.user));
        location.reload();
    } else {
        alert(res.errorMessage || 'Ошибка входа');
    }
}

async function registerUser() {
    const email = document.getElementById('loginEmail').value;
    const password = document.getElementById('loginPass').value;
    const username = email.split('@') + '_' + Math.floor(Math.random() * 100);
    const res = await apiFetch('/api/auth/register', { method: 'POST', body: JSON.stringify({ username, email, password }) });
    if (res.isSuccess) {
        alert('Пользователь ' + res.result.username + ' зарегистрирован с ролью USER! Теперь нажмите «Войти».');
    } else {
        alert(res.errorMessage || 'Ошибка регистрации');
    }
}

function logoutUser() {
    localStorage.removeItem('jwt_token');
    localStorage.removeItem('jwt_user');
    location.reload();
}

document.addEventListener('DOMContentLoaded', renderAuthBar);
