package handler

const HTMLTemplate = `<!DOCTYPE html>
<html lang="pt-BR">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Chat em Tempo Real</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            height: 100vh;
            display: flex;
            justify-content: center;
            align-items: center;
        }
        .container {
            width: 90%;
            max-width: 1000px;
            height: 90vh;
            background: white;
            border-radius: 20px;
            box-shadow: 0 20px 60px rgba(0,0,0,0.3);
            display: flex;
            overflow: hidden;
        }
        .sidebar {
            width: 250px;
            background: #2d3748;
            color: white;
            padding: 20px;
            display: flex;
            flex-direction: column;
        }
        .sidebar h2 { margin-bottom: 20px; font-size: 1.2rem; }
        .room-list { flex: 1; overflow-y: auto; }
        .room-item {
            padding: 12px;
            margin-bottom: 8px;
            background: #4a5568;
            border-radius: 8px;
            cursor: pointer;
            transition: all 0.3s;
        }
        .room-item:hover, .room-item.active {
            background: #667eea;
        }
        .room-item small { display: block; opacity: 0.7; font-size: 0.8rem; }
        .main {
            flex: 1;
            display: flex;
            flex-direction: column;
        }
        .header {
            background: #f7fafc;
            padding: 20px;
            border-bottom: 1px solid #e2e8f0;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        .messages {
            flex: 1;
            overflow-y: auto;
            padding: 20px;
            background: #f7fafc;
        }
        .message {
            margin-bottom: 15px;
            animation: fadeIn 0.3s;
        }
        @keyframes fadeIn {
            from { opacity: 0; transform: translateY(10px); }
            to { opacity: 1; transform: translateY(0); }
        }
        .message-header {
            display: flex;
            align-items: center;
            margin-bottom: 5px;
        }
        .username {
            font-weight: bold;
            color: #2d3748;
            margin-right: 10px;
        }
        .time {
            font-size: 0.75rem;
            color: #718096;
        }
        .message-content {
            background: white;
            padding: 12px;
            border-radius: 12px;
            display: inline-block;
            max-width: 70%;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .message.system .message-content {
            background: #e6fffa;
            color: #234e52;
            font-style: italic;
        }
        .message.own .message-content {
            background: #667eea;
            color: white;
        }
        .input-area {
            padding: 20px;
            background: white;
            border-top: 1px solid #e2e8f0;
            display: flex;
            gap: 10px;
        }
        input[type="text"] {
            flex: 1;
            padding: 12px;
            border: 2px solid #e2e8f0;
            border-radius: 25px;
            outline: none;
            font-size: 1rem;
        }
        input[type="text"]:focus {
            border-color: #667eea;
        }
        button {
            padding: 12px 24px;
            background: #667eea;
            color: white;
            border: none;
            border-radius: 25px;
            cursor: pointer;
            font-size: 1rem;
            transition: background 0.3s;
        }
        button:hover { background: #5568d3; }
        .login-modal {
            position: fixed;
            top: 0; left: 0; right: 0; bottom: 0;
            background: rgba(0,0,0,0.8);
            display: flex;
            justify-content: center;
            align-items: center;
            z-index: 1000;
        }
        .login-box {
            background: white;
            padding: 40px;
            border-radius: 20px;
            width: 90%;
            max-width: 400px;
        }
        .login-box h2 { margin-bottom: 20px; color: #2d3748; }
        .login-box input {
            width: 100%;
            margin-bottom: 15px;
        }
        .hidden { display: none !important; }
        .user-count {
            background: #48bb78;
            color: white;
            padding: 4px 12px;
            border-radius: 20px;
            font-size: 0.875rem;
        }
    </style>
</head>
<body>
    <div class="login-modal" id="loginModal">
        <div class="login-box">
            <h2>🚀 Entrar no Chat</h2>
            <input type="text" id="usernameInput" placeholder="Seu nome de usuário" maxlength="20">
            <input type="text" id="roomInput" placeholder="ID da sala (opcional)" value="general">
            <button onclick="connect()">Conectar</button>
        </div>
    </div>

    <div class="container hidden" id="chatContainer">
        <div class="sidebar">
            <h2>💬 Salas</h2>
            <div class="room-list" id="roomList"></div>
            <div style="margin-top: auto;">
                <input type="text" id="newRoomName" placeholder="Nova sala" style="width: 100%; margin-bottom: 8px;">
                <button onclick="createRoom()" style="width: 100%;">Criar Sala</button>
            </div>
        </div>
        <div class="main">
            <div class="header">
                <div>
                    <h3 id="currentRoomName">Geral</h3>
                    <small id="currentRoomDesc">Sala de bate-papo geral</small>
                </div>
                <span class="user-count" id="userCount">0 online</span>
            </div>
            <div class="messages" id="messages"></div>
            <div class="input-area">
                <input type="text" id="messageInput" placeholder="Digite sua mensagem..." maxlength="500" onkeypress="if(event.key==='Enter')sendMessage()">
                <button onclick="sendMessage()">Enviar</button>
            </div>
        </div>
    </div>

    <script>
        let ws;
        let username;
        let currentRoom = 'general';
        let reconnectInterval;

        async function connect() {
            username = document.getElementById('usernameInput').value.trim();
            const room = document.getElementById('roomInput').value.trim() || 'general';
            
            if (!username) {
                alert('Por favor, digite um nome de usuário');
                return;
            }

            currentRoom = room;
            
            // Carregar salas disponíveis
            await loadRooms();
            
            // Conectar WebSocket
            connectWebSocket();
            
            // Carregar histórico
            await loadHistory();
            
            document.getElementById('loginModal').classList.add('hidden');
            document.getElementById('chatContainer').classList.remove('hidden');
            document.getElementById('currentRoomName').textContent = currentRoom;
        }

        function connectWebSocket() {
            const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
            ws = new WebSocket(protocol + '//' + window.location.host + '/ws?room=' + currentRoom + '&username=' + encodeURIComponent(username));

            ws.onopen = () => {
                console.log('Conectado!');
                clearInterval(reconnectInterval);
            };

            ws.onmessage = (event) => {
                const msg = JSON.parse(event.data);
                displayMessage(msg);
            };

            ws.onclose = () => {
                console.log('Desconectado, tentando reconectar...');
                reconnectInterval = setTimeout(connectWebSocket, 3000);
            };

            ws.onerror = (error) => {
                console.error('WebSocket error:', error);
            };
        }

        async function loadRooms() {
            try {
                const response = await fetch('/api/rooms');
                const rooms = await response.json();
                const roomList = document.getElementById('roomList');
                roomList.innerHTML = '';
                
                rooms.forEach(room => {
                    const div = document.createElement('div');
                    div.className = 'room-item' + (room.id === currentRoom ? ' active' : '');
                    div.innerHTML = '<strong>' + room.name + '</strong><small>' + room.user_count + ' online • ' + room.description + '</small>';
                    div.onclick = () => switchRoom(room.id);
                    roomList.appendChild(div);
                });
            } catch (err) {
                console.error('Erro ao carregar salas:', err);
            }
        }

        async function loadHistory() {
            try {
                const response = await fetch('/api/messages?room=' + currentRoom);
                const data = await response.json();
                document.getElementById('messages').innerHTML = '';
                data.messages.forEach(msg => displayMessage(msg));
            } catch (err) {
                console.error('Erro ao carregar histórico:', err);
            }
        }

        function displayMessage(msg) {
            const messagesDiv = document.getElementById('messages');
            const messageDiv = document.createElement('div');
            let className = 'message';
            if (msg.username === 'Sistema') className += ' system';
            if (msg.username === username) className += ' own';
            messageDiv.className = className;
            
            const time = new Date(msg.created_at).toLocaleTimeString('pt-BR', {hour: '2-digit', minute:'2-digit'});
            
            messageDiv.innerHTML = '<div class="message-header"><span class="username">' + escapeHtml(msg.username) + '</span><span class="time">' + time + '</span></div><div class="message-content">' + escapeHtml(msg.content) + '</div>';
            
            messagesDiv.appendChild(messageDiv);
            messagesDiv.scrollTop = messagesDiv.scrollHeight;
        }

        function sendMessage() {
            const input = document.getElementById('messageInput');
            const content = input.value.trim();
            
            if (!content || !ws || ws.readyState !== WebSocket.OPEN) return;
            
            ws.send(JSON.stringify({content: content}));
            input.value = '';
        }

        async function createRoom() {
            const name = document.getElementById('newRoomName').value.trim();
            if (!name) return;
            
            const id = name.toLowerCase().replace(/\s+/g, '-');
            
            try {
                await fetch('/api/rooms', {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify({id: id, name: name, description: 'Sala criada pelo usuário'})
                });
                
                document.getElementById('newRoomName').value = '';
                await loadRooms();
            } catch (err) {
                alert('Erro ao criar sala');
            }
        }

        function switchRoom(roomId) {
            if (roomId === currentRoom) return;
            currentRoom = roomId;
            if (ws) ws.close();
            connectWebSocket();
            loadHistory();
            loadRooms();
        }

        function escapeHtml(text) {
            const div = document.createElement('div');
            div.textContent = text;
            return div.innerHTML;
        }

        // Atualizar lista de salas a cada 5 segundos
        setInterval(loadRooms, 5000);
    </script>
</body>
</html>`
