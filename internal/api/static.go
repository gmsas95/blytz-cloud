package api

const indexHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Blytz - Your Personal AI Assistant</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            color: white;
        }
        .container {
            text-align: center;
            max-width: 600px;
            padding: 2rem;
        }
        h1 { font-size: 3rem; margin-bottom: 1rem; }
        p { font-size: 1.25rem; margin-bottom: 2rem; opacity: 0.9; }
        .cta-form {
            display: flex;
            gap: 1rem;
            justify-content: center;
            flex-wrap: wrap;
        }
        input[type="email"] {
            padding: 1rem 1.5rem;
            border: none;
            border-radius: 50px;
            font-size: 1rem;
            width: 300px;
            outline: none;
        }
        button {
            padding: 1rem 2rem;
            background: #ff6b6b;
            color: white;
            border: none;
            border-radius: 50px;
            font-size: 1rem;
            cursor: pointer;
            transition: transform 0.2s;
        }
        button:hover { transform: scale(1.05); }
        .price { margin-top: 2rem; font-size: 1.5rem; font-weight: bold; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Your Personal AI Assistant</h1>
        <p>Choose your agent. Pick your AI model. Deploy in seconds.</p>
        <form class="cta-form" onsubmit="event.preventDefault(); window.location.href='/configure?email=' + encodeURIComponent(this.email.value);">
            <input type="email" name="email" placeholder="Enter your email" required>
            <button type="submit">Get Started →</button>
        </form>
        <div class="price">$29/month</div>
    </div>
</body>
</html>`

const configureHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Configure Your Assistant - Blytz</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: #f5f5f5;
            min-height: 100vh;
            padding: 2rem;
        }
        .container {
            max-width: 600px;
            margin: 0 auto;
            background: white;
            padding: 2rem;
            border-radius: 16px;
            box-shadow: 0 4px 20px rgba(0,0,0,0.1);
        }
        h1 { color: #333; margin-bottom: 2rem; text-align: center; }
        .step { margin-bottom: 2rem; }
        .step-number {
            display: inline-block;
            width: 32px;
            height: 32px;
            background: #667eea;
            color: white;
            border-radius: 50%;
            text-align: center;
            line-height: 32px;
            margin-right: 0.5rem;
            font-weight: bold;
        }
        label { display: block; margin-bottom: 0.5rem; color: #555; font-weight: 500; }
        input, textarea, select {
            width: 100%;
            padding: 0.75rem;
            border: 2px solid #e0e0e0;
            border-radius: 8px;
            font-size: 1rem;
            font-family: inherit;
            transition: border-color 0.2s;
        }
        input:focus, textarea:focus, select:focus {
            outline: none;
            border-color: #667eea;
        }
        textarea { min-height: 150px; resize: vertical; }
        select { background: white; cursor: pointer; }
        .hint {
            font-size: 0.875rem;
            color: #888;
            margin-top: 0.25rem;
        }
        button {
            width: 100%;
            padding: 1rem;
            background: #667eea;
            color: white;
            border: none;
            border-radius: 8px;
            font-size: 1.1rem;
            cursor: pointer;
            transition: background 0.2s;
        }
        button:hover { background: #5568d3; }
        button:disabled { background: #ccc; cursor: not-allowed; }
        .error { color: #ff6b6b; margin-top: 1rem; text-align: center; }
        .loading { display: none; text-align: center; margin-top: 1rem; }
        .agent-option {
            border: 2px solid #e0e0e0;
            border-radius: 8px;
            padding: 1rem;
            margin-bottom: 0.5rem;
            cursor: pointer;
            transition: all 0.2s;
        }
        .agent-option:hover { border-color: #667eea; }
        .agent-option.selected {
            border-color: #667eea;
            background: #f0f4ff;
        }
        .agent-name { font-weight: bold; color: #333; }
        .agent-desc { font-size: 0.875rem; color: #666; margin-top: 0.25rem; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Configure Your Assistant</h1>
        <form id="configForm">
            <div class="step">
                <label><span class="step-number">1</span>What should I call you?</label>
                <input type="text" id="assistantName" placeholder="e.g., Alex, Mike, Assistant" required>
            </div>
            
            <div class="step">
                <label><span class="step-number">2</span>What do you want help with?</label>
                <textarea id="customInstructions" placeholder="I'm a freelance developer. I need help with:
- Drafting proposals for new clients
- Researching competitors and technologies
- Following up on outstanding invoices" required></textarea>
                <div class="hint">Be specific. The more detail, the better your assistant will be.</div>
            </div>
            
            <div class="step">
                <label><span class="step-number">3</span>Choose Your Agent</label>
                <div id="agentOptions">
                    <div class="loading">Loading available agents...</div>
                </div>
                <input type="hidden" id="agentTypeId" value="openclaw">
            </div>
            
            <div class="step">
                <label><span class="step-number">4</span>Choose Your AI Model</label>
                <select id="llmProviderId" required>
                    <option value="">Loading providers...</option>
                </select>
            </div>
            
            <div class="step">
                <label><span class="step-number">5</span>Your API Key</label>
                <input type="password" id="llmApiKey" placeholder="sk-..." required>
                <div class="hint" id="apiKeyHint">Enter your API key from the AI provider</div>
            </div>
            
            <div class="step">
                <label><span class="step-number">6</span>Telegram Bot Token</label>
                <input type="text" id="telegramToken" placeholder="123456789:ABCdefGHIjklMNOpqrsTUVwxyz" required>
                <div class="hint">Get one free from @BotFather on Telegram</div>
            </div>
            
            <button type="submit" id="submitBtn">Continue to Payment →</button>
            <div id="error" class="error"></div>
            <div id="loading" class="loading">Setting up your assistant...</div>
        </form>
    </div>
    <script>
        const urlParams = new URLSearchParams(window.location.search);
        const email = urlParams.get('email') || '';
        
        // Load marketplace data
        async function loadMarketplace() {
            try {
                // Load agents
                const agentsRes = await fetch('/api/marketplace/agents');
                const agentsData = await agentsRes.json();
                renderAgents(agentsData.agents);
                
                // Load LLM providers
                const llmRes = await fetch('/api/marketplace/llm-providers');
                const llmData = await llmRes.json();
                renderProviders(llmData.providers);
            } catch (err) {
                console.error('Failed to load marketplace:', err);
                document.getElementById('error').textContent = 'Failed to load options. Please refresh.';
            }
        }
        
        function renderAgents(agents) {
            const container = document.getElementById('agentOptions');
            container.innerHTML = '';
            
            agents.forEach(function(agent) {
                const div = document.createElement('div');
                div.className = 'agent-option';
                div.innerHTML = '<div class="agent-name">' + agent.name + '</div>' +
                               '<div class="agent-desc">' + agent.description + '</div>';
                
                div.addEventListener('click', function() {
                    container.querySelectorAll('.agent-option').forEach(function(el) {
                        el.classList.remove('selected');
                    });
                    div.classList.add('selected');
                    document.getElementById('agentTypeId').value = agent.id;
                });
                
                container.appendChild(div);
                
                // Select first by default
                if (container.children.length === 1) {
                    div.click();
                }
            });
        }
        
        function renderProviders(providers) {
            const select = document.getElementById('llmProviderId');
            select.innerHTML = '<option value="">Select AI provider...</option>';
            
            providers.forEach(function(provider) {
                const option = document.createElement('option');
                option.value = provider.id;
                option.textContent = provider.name;
                select.appendChild(option);
            });
            
            // Update hint when selection changes
            select.addEventListener('change', function() {
                const provider = providers.find(function(p) { return p.id === select.value; });
                if (provider) {
                    const hint = document.getElementById('apiKeyHint');
                    const input = document.getElementById('llmApiKey');
                    if (provider.id === 'ollama') {
                        hint.textContent = 'Ollama runs locally - no API key needed for local models';
                        input.required = false;
                        input.placeholder = 'Optional';
                    } else {
                        hint.textContent = 'Enter your ' + provider.name + ' API key';
                        input.required = true;
                        input.placeholder = 'sk-...';
                    }
                }
            });
            
            // Select first option
            if (providers.length > 0) {
                select.value = providers[0].id;
                select.dispatchEvent(new Event('change'));
            }
        }
        
        document.getElementById('configForm').addEventListener('submit', async function(e) {
            e.preventDefault();
            document.getElementById('error').textContent = '';
            document.getElementById('loading').style.display = 'block';
            document.getElementById('submitBtn').disabled = true;
            
            const data = {
                email: email,
                assistant_name: document.getElementById('assistantName').value,
                custom_instructions: document.getElementById('customInstructions').value,
                telegram_bot_token: document.getElementById('telegramToken').value,
                agent_type_id: document.getElementById('agentTypeId').value,
                llm_provider_id: document.getElementById('llmProviderId').value,
                llm_api_key: document.getElementById('llmApiKey').value
            };
            
            try {
                const response = await fetch('/api/signup', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(data)
                });
                
                const result = await response.json();
                
                if (response.ok) {
                    window.location.href = result.checkout_url;
                } else {
                    document.getElementById('error').textContent = result.message || 'Something went wrong';
                    document.getElementById('loading').style.display = 'none';
                    document.getElementById('submitBtn').disabled = false;
                }
            } catch (err) {
                document.getElementById('error').textContent = 'Network error. Please try again.';
                document.getElementById('loading').style.display = 'none';
                document.getElementById('submitBtn').disabled = false;
            }
        });
        
        loadMarketplace();
    </script>
</body>
</html>`

const successHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Success! - Blytz</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            color: white;
        }
        .container {
            text-align: center;
            max-width: 600px;
            padding: 2rem;
        }
        .checkmark {
            font-size: 4rem;
            margin-bottom: 1rem;
        }
        h1 { font-size: 2.5rem; margin-bottom: 1rem; }
        p { font-size: 1.25rem; margin-bottom: 2rem; opacity: 0.9; }
        .button {
            display: inline-block;
            padding: 1rem 2rem;
            background: white;
            color: #667eea;
            text-decoration: none;
            border-radius: 50px;
            font-weight: bold;
            transition: transform 0.2s;
        }
        .button:hover { transform: scale(1.05); }
        .tips {
            margin-top: 3rem;
            text-align: left;
            background: rgba(255,255,255,0.1);
            padding: 1.5rem;
            border-radius: 12px;
        }
        .tips h3 { margin-bottom: 1rem; }
        .tips ul { margin-left: 1.5rem; }
        .tips li { margin-bottom: 0.5rem; }
    </style>
</head>
<body>
    <div class="container">
        <div class="checkmark">🎉</div>
        <h1>Your Assistant is Ready!</h1>
        <p>Payment confirmed. Your AI assistant is being deployed now.</p>
        <a id="telegramLink" href="#" class="button" target="_blank">Open in Telegram</a>
        
        <div class="tips">
            <h3>Quick Tips:</h3>
            <ul>
                <li>Just start chatting - your assistant already knows your context</li>
                <li>It remembers conversations and learns over time</li>
                <li>Cancel anytime from your dashboard</li>
            </ul>
        </div>
    </div>
    <script>
        const urlParams = new URLSearchParams(window.location.search);
        const customerId = urlParams.get('customer_id');
        
        if (customerId) {
            fetch('/api/status/' + customerId)
                .then(function(res) { return res.json(); })
                .then(function(data) {
                    if (data.telegram_bot_username) {
                        document.getElementById('telegramLink').href = 
                            'https://t.me/' + data.telegram_bot_username;
                    }
                });
        }
    </script>
</body>
</html>`
