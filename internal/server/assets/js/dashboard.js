let socket;
    let reconnectAttempts = 0;
    const maxReconnectAttempts = 5;
    const reconnectDelay = 3000;

    function connectWebSocket() {
        socket = new WebSocket('ws://' + window.location.host + '/ws');

        socket.onopen = function() {
            console.log('WebSocket connected');
            reconnectAttempts = 0;
        };

        socket.onmessage = function(event) {
            const data = JSON.parse(event.data);
            updateDashboard(data);
        };

        socket.onclose = function() {
            console.log('WebSocket disconnected');
            if (reconnectAttempts < maxReconnectAttempts) {
                setTimeout(connectWebSocket, reconnectDelay);
                reconnectAttempts++;
            } else {
                console.error('Max reconnect attempts reached');
            }
        };
    }

    function fetchInitialData() {
        fetch('/initial-data')
            .then(response => response.json())
            .then(data => {
                updateDashboard(data);
                document.getElementById('loading').style.display = 'none';
                document.getElementById('dashboard').style.display = 'block';
            })
            .catch(error => console.error('Error fetching initial data:', error));
    }

    function updateDashboard(data) {
        const versionElement = document.getElementById('version');
        if (versionElement) {
            versionElement.textContent = data.Version;
            if (data.Version === "dev") {
                versionElement.textContent = "Development Version";
                versionElement.style.backgroundColor = "#dc3545";
            }
        }

        const container = document.getElementById('characters-container');
        if (!container) return;

        if (Object.keys(data.Status).length === 0) {
            container.innerHTML = '<article><p>No characters found, start adding a new character.</p></article>';
            return;
        }

        for (const [key, value] of Object.entries(data.Status)) {
            let card = document.getElementById(`card-${key}`);
            if (!card) {
                card = createCharacterCard(key);
                container.appendChild(card);
            }
            updateCharacterCard(card, key, value, data.DropCount[key]);
        }

        // Remove cards for characters that no longer exist
        Array.from(container.children).forEach(card => {
            if (!data.Status.hasOwnProperty(card.id.replace('card-', ''))) {
                container.removeChild(card);
            }
        });
    }


    function createCharacterCard(key) {
        const card = document.createElement('div');
        card.className = 'character-card';
        card.id = `card-${key}`;

        card.innerHTML = `
            <div class="character-header">
                <div class="character-name">
                    <span>${key}</span>
                     <div class="status-indicator"></div>
                </div>
                <div class="character-controls">
                    <button class="btn btn-outline companion-join-btn" onclick="showCompanionJoinPopup('${key}')" style="display:none;">
                        <i class="bi bi-door-open btn-icon"></i>Join Game
                    </button>
                    <button class="btn btn-outline" onclick="location.href='/debug?characterName=${key}'">
                        <i class="bi bi-bug btn-icon"></i>Debug
                    </button>
                    <button class="btn btn-outline" onclick="location.href='/supervisorSettings?supervisor=${key}'">
                        <i class="bi bi-gear btn-icon"></i>Settings
                    </button>
                    <button class="start-pause btn btn-start" data-character="${key}">
                        <i class="bi bi-play-fill btn-icon"></i>Start
                    </button>
                    <button class="stop btn btn-stop" data-character="${key}" style="display:none;">
                        <i class="bi bi-stop-fill btn-icon"></i>Stop
                    </button>
                    <button class="btn btn-outline attach-btn" onclick="showAttachPopup('${key}')" style="display:none;">
                        <i class="bi bi-link-45deg btn-icon"></i>Attach
                    </button>
                    <button class="toggle-details">
                        <i class="bi bi-chevron-down"></i>
                    </button>
                </div>
            </div>
            <div class="character-details">
                <div class="status-details">
                    <span class="status-badge"></span>
                </div>
                <div class="stats-grid">
                    <div class="stat-item">
                        <div class="stat-label">Games</div>
                        <div class="stat-value runs">0</div>
                    </div>
                    <div class="stat-item">
                        <div class="stat-label">Drops</div>
                        <div class="stat-value drops">None</div>
                    </div>
                    <div class="stat-item">
                        <div class="stat-label">Chickens</div>
                        <div class="stat-value chickens">0</div>
                    </div>
                    <div class="stat-item">
                        <div class="stat-label">Deaths</div>
                        <div class="stat-value deaths">0</div>
                    </div>
                    <div class="stat-item">
                        <div class="stat-label">Errors</div>
                        <div class="stat-value errors">0</div>
                    </div>
                </div>
                <div class="run-stats"></div>
            </div>
        `;

        setupEventListeners(card, key);
        return card;
    }


    function setupEventListeners(card, key) {
        if (!card) return;

        const toggleDetailsBtn = card.querySelector('.toggle-details');
        const startPauseBtn = card.querySelector('.start-pause');
        const stopBtn = card.querySelector('.stop');

        if (toggleDetailsBtn) {
            toggleDetailsBtn.addEventListener('click', function() {
                card.classList.toggle('expanded');
                this.querySelector('i').style.transform = card.classList.contains('expanded') ? 'rotate(180deg)' : 'rotate(0deg)';
                saveExpandedState();
            });
        }

        if (startPauseBtn) {
            startPauseBtn.addEventListener('click', function() {
                const currentStatus = this.className.includes('btn-start') ? 'Not Started' :
                    this.className.includes('btn-pause') ? 'In game' :
                        'Paused';
                let action;
                if (currentStatus === 'Not Started') {
                    action = 'start';
                } else if (currentStatus === 'In game') {
                    action = 'togglePause';
                } else { // Paused
                    action = 'togglePause';
                }
                fetch(`/${action}?characterName=${key}`)
                    .then(response => response.json())
                    .then(data => {
                        updateDashboard(data);
                    })
                    .catch(error => console.error('Error:', error));
            });
        }
        if (stopBtn) {
            stopBtn.addEventListener('click', function() {
                fetch(`/stop?characterName=${key}`).then(() => fetchInitialData());
            });
        }
    }


    function updateStatusPosition(card, isExpanded) {
        if (!card) return;

        const statusBadge = card.querySelector('.status-badge');
        const headerStatusContainer = card.querySelector('.character-name-status');
        const detailsStatusContainer = card.querySelector('.status-details');

        if (!statusBadge || !headerStatusContainer || !detailsStatusContainer) return;

        if (isExpanded) {
            detailsStatusContainer.insertBefore(statusBadge, detailsStatusContainer.firstChild);
        } else {
            headerStatusContainer.appendChild(statusBadge);
        }
    }

    function updateCharacterCard(card, key, value, dropCount) {
        if (!card) return;

        const startPauseBtn = card.querySelector('.start-pause');
        const stopBtn = card.querySelector('.stop');
        const attachBtn = card.querySelector('.attach-btn');
        const companionJoinBtn = card.querySelector('.companion-join-btn');
        const statusDetails = card.querySelector('.status-details');
        const statusBadge = statusDetails.querySelector('.status-badge');
        const statusIndicator = card.querySelector('.status-indicator');

        if (statusBadge && statusDetails) {
            updateStatus(statusBadge, statusDetails, value.SupervisorStatus);
        }
        
        if (statusIndicator) {
                updateStatusIndicator(statusIndicator, value.SupervisorStatus);
        }

        if (startPauseBtn && stopBtn && attachBtn) {
            updateButtons(startPauseBtn, stopBtn, attachBtn, value.SupervisorStatus);
        }

        // Update companion join button visibility
        if (companionJoinBtn) {
            const isCompanionFollower = value.IsCompanionFollower || false;
            // Only show the button if it's a companion follower AND the supervisor is running
            const isRunning = value.SupervisorStatus === "In game" || value.SupervisorStatus === "Paused" || value.SupervisorStatus === "Starting";
            companionJoinBtn.style.display = (isCompanionFollower && isRunning) ? 'inline-flex' : 'none';
        }

        updateStats(card, key, value.Games, dropCount);
        updateRunStats(card, value.Games);
        
        if (statusDetails) {
            updateStartedTime(statusDetails, value.StartedAt);
        }
    }

    function updateStatusIndicator(statusIndicator, status) {
        statusIndicator.classList.remove('in-game', 'paused', 'stopped');
        if (status === "In game") {
            statusIndicator.classList.add('in-game');
        } else if (status === "Starting") {
            statusIndicator.classList.add('paused');
        } else if (status === "Paused") {
            statusIndicator.classList.add('paused');
        } else {
            statusIndicator.classList.add('stopped');
        }
    }

    function updateStatus(statusBadge, statusDetails, status) {
        if (!statusBadge || !statusDetails) return;

        const statusText = status || 'Not started';
        statusBadge.innerHTML = `<span class="status-label">Status:</span> <span class="status-value">${statusText}</span>`;
        statusBadge.className = `status-badge status-${statusText.toLowerCase().replace(' ', '')}`;
    }

    function updateStartedTime(statusDetails, startedAt) {
        const startTime = new Date(startedAt);
        const now = new Date();
        
        let runningForElement = statusDetails.querySelector('.running-for');
        if (!runningForElement) {
            runningForElement = document.createElement('div');
            runningForElement.className = 'running-for';
            statusDetails.appendChild(runningForElement);
        }
        
        if (startTime.getFullYear() === 1) {
            runningForElement.textContent = 'Running for: N/A';
            return;
        }
        
        const timeDiff = now - startTime;
        if (timeDiff < 0) {
            runningForElement.textContent = 'Running for: N/A';
            return;
        }
        
        const duration = formatDuration(timeDiff);
        runningForElement.textContent = `Running for: ${duration}`;
    }

function updateButtons(startPauseBtn, stopBtn, attachBtn, status) {
    if (status === "Paused") {
        startPauseBtn.innerHTML = '<i class="bi bi-play-fill btn-icon"></i>Resume';
        startPauseBtn.className = 'start-pause btn btn-resume';
        stopBtn.style.display = 'inline-block';
        attachBtn.style.display = 'none';
    } else if (status === "In game" || status === "Starting") {
        startPauseBtn.innerHTML = '<i class="bi bi-pause-fill btn-icon"></i>Pause';
        startPauseBtn.className = 'start-pause btn btn-pause';
        stopBtn.style.display = 'inline-block';
        attachBtn.style.display = 'none';
    } else {
        startPauseBtn.innerHTML = '<i class="bi bi-play-fill btn-icon"></i>Start';
        startPauseBtn.className = 'start-pause btn btn-start';
        stopBtn.style.display = 'none';
        attachBtn.style.display = 'inline-block';
    }
}

    function updateStats(card, key, games, dropCount) {
        const stats = calculateStats(games);
        
        card.querySelector('.runs').textContent = stats.totalGames;
        card.querySelector('.drops').innerHTML = dropCount === undefined ? 'None' : 
            (dropCount === 0 ? 'None' : `<a href="/drops?supervisor=${key}">${dropCount}</a>`);
        card.querySelector('.chickens').textContent = stats.totalChickens;
        card.querySelector('.deaths').textContent = stats.totalDeaths;
        card.querySelector('.errors').textContent = stats.totalErrors;
    }


    function updateRunStats(card, games) {
    const runStats = calculateRunStats(games);
    const runStatsElement = card.querySelector('.run-stats');
    runStatsElement.innerHTML = '<h3>Run Statistics</h3>';

    if (Object.keys(runStats).length === 0) {
        runStatsElement.innerHTML += '<p>No run data available yet.</p>';
        return;
    }

    const runStatsGrid = document.createElement('div');
    runStatsGrid.className = 'run-stats-grid';

    for (const [runName, stats] of Object.entries(runStats)) {
        const runElement = document.createElement('div');
        runElement.className = 'run-stat';
        if (stats.isCurrentRun) {
            runElement.classList.add('current-run');
        }
        runElement.innerHTML = `
            <h4>${runName}${stats.isCurrentRun ? ' <span class="current-run-indicator">Current</span>' : ''}</h4>
            <div class="run-stat-content">
                <div class="run-stat-item" title="Fastest Run">
                    <span class="stat-label">Fastest:</span> ${formatDuration(stats.shortestTime)}
                </div>
                <div class="run-stat-item" title="Slowest Run">
                    <span class="stat-label">Slowest:</span> ${formatDuration(stats.longestTime)}
                </div>
                <div class="run-stat-item" title="Average Run">
                    <span class="stat-label">Average:</span> ${formatDuration(stats.averageTime)}
                </div>
                <div class="run-stat-item" title="Total Runs">
                    <span class="stat-label">Total:</span> ${stats.runCount}
                </div>
                <div class="run-stat-item" title="Errors">
                    <span class="stat-label">Errors:</span> ${stats.errorCount}
                </div>
                <div class="run-stat-item" title="Chickens">
                    <span class="stat-label">Chickens:</span> ${stats.runChickens}
                </div>
                <div class="run-stat-item" title="Deaths">
                    <span class="stat-label">Deaths:</span> ${stats.runDeaths}
                </div>
            </div>
        `;
        runStatsGrid.appendChild(runElement);
    }

        runStatsElement.appendChild(runStatsGrid);
    }   


    function calculateRunStats(games) {
        if (!games || games.length === 0) {
            return {};
        }

        const runStats = {};

        games.forEach(game => {
            if (game.Runs && Array.isArray(game.Runs)) {
                game.Runs.forEach(run => {
                    if (!runStats[run.Name]) {
                        runStats[run.Name] = { 
                            shortestTime: Infinity, 
                            longestTime: 0, 
                            totalTime: 0,
                            errorCount: 0, 
                            runCount: 0,
                            runChickens: 0,
                            runDeaths: 0,
                            successfulRunCount: 0,
                            isCurrentRun: false
                        };
                    }

                    // Check if this is the current run
                    if (run.Reason === "") {
                        runStats[run.Name].isCurrentRun = true;
                    }

                    const runTime = new Date(run.FinishedAt) - new Date(run.StartedAt);
                    if (run.FinishedAt !== "0001-01-01T00:00:00Z" && runTime > 0) {
                        runStats[run.Name].runCount++;

                        if (run.Reason === 'ok') {
                            runStats[run.Name].shortestTime = Math.min(runStats[run.Name].shortestTime, runTime);
                            runStats[run.Name].longestTime = Math.max(runStats[run.Name].longestTime, runTime);
                            runStats[run.Name].totalTime += runTime;
                            runStats[run.Name].successfulRunCount++;
                        }
                    }

                    if (run.Reason == 'error') {
                        runStats[run.Name].errorCount++;
                    }

                    if (run.Reason == 'chicken') {
                        runStats[run.Name].runChickens++;
                    }

                    if (run.Reason == 'death') {
                        runStats[run.Name].runDeaths++;
                    }
                });
            }
        });

        // Calculate average time for successful runs
        for (const stats of Object.values(runStats)) {
            if (stats.successfulRunCount > 0) {
                stats.averageTime = stats.totalTime / stats.successfulRunCount;
            } else {
                stats.shortestTime = 0;
                stats.longestTime = 0;
                stats.averageTime = 0;
            }
        }

        return runStats;
    }

    function calculateStats(games) {
        if (!games || games.length === 0) {
            return { totalGames: 0, totalChickens: 0, totalDeaths: 0, totalErrors: 0 };
        }

        return games.reduce((acc, game) => {
            acc.totalGames++;
            if (game.Reason === 'chicken') acc.totalChickens++;
            else if (game.Reason === 'death') acc.totalDeaths++;
            else if (game.Reason === 'error') acc.totalErrors++;
            return acc;
        }, { totalGames: 0, totalChickens: 0, totalDeaths: 0, totalErrors: 0 });
    } 

    function formatDuration(ms) {
        if (!isFinite(ms) || ms < 0) {
            return 'N/A';
        }
        const seconds = Math.floor(ms / 1000);
        const minutes = Math.floor(seconds / 60);
        const hours = Math.floor(minutes / 60);
        const days = Math.floor(hours / 24);

        if (days > 0) return `${days}d ${hours % 24}h`;
        if (hours > 0) return `${hours}h ${minutes % 60}m`;
        if (minutes > 0) return `${minutes}m ${seconds % 60}s`;
        return `${seconds}s`;
    }

    function saveExpandedState() {
        const expandedCards = Array.from(document.querySelectorAll('.character-card.expanded'))
            .map(card => card.id);
        localStorage.setItem('expandedCards', JSON.stringify(expandedCards));
    }

    function restoreExpandedState() {
        const expandedCards = JSON.parse(localStorage.getItem('expandedCards')) || [];
        expandedCards.forEach(cardId => {
            const card = document.getElementById(cardId);
            if (card) {
                card.classList.add('expanded');
                const toggleBtn = card.querySelector('.toggle-details i');
                if (toggleBtn) {
                    toggleBtn.style.transform = 'rotate(180deg)';
                }
            }
        });
    }

    function showAttachPopup(characterName) {
        const popup = document.createElement('div');
        popup.className = 'attach-popup';
        popup.innerHTML = `
            <h3>Attach to Process</h3>
            <div id="popup-content">
                <div class="loading-spinner"></div>
                <p>Loading processes...</p>
            </div>
        `;
    
        document.body.appendChild(popup);
    
        // Fetch and populate the process list
        fetchProcessList(characterName);
    }

    function fetchProcessList(characterName) {
        fetch('/process-list')
            .then(response => response.json())
            .then(processes => {
                const popup = document.querySelector('.attach-popup');
                
                if (!processes || processes.length === 0) {
                    popup.innerHTML = `
                        <h3>No D2R Processes Found</h3>
                        <p>There are no Diablo II: Resurrected processes currently running.</p>
                        <button onclick="closeAttachPopup()" class="btn btn-primary">Close</button>
                    `;
                } else {
                    popup.innerHTML = `
                        <h3>Attach to Process</h3>
                        <input type="text" id="process-search" placeholder="Search processes...">
                        <table>
                            <thead>
                                <tr>
                                    <th>Window Title</th>
                                    <th>Process Name</th>
                                    <th>PID</th>
                                </tr>
                            </thead>
                            <tbody id="process-list-body"></tbody>
                        </table>
                        <div class="selected-process">
                            <span>Selected Process: </span>
                            <span id="selected-pid">None</span>
                        </div>
                        <div class="popup-buttons">
                            <button id="choose-process" class="btn btn-primary" disabled>Attach</button>
                            <button id="cancel-attach" class="btn">Cancel</button>
                        </div>
                    `;
    
                    const tbody = document.getElementById('process-list-body');
                    processes.forEach(process => {
                        const row = document.createElement('tr');
                        row.innerHTML = `
                            <td>${process.windowTitle}</td>
                            <td>${process.processName}</td>
                            <td>${process.pid}</td>
                        `;
                        row.addEventListener('click', () => selectProcess(row, process.pid));
                        tbody.appendChild(row);
                    });
    
                    // Add event listeners
                    document.getElementById('choose-process').addEventListener('click', () => chooseProcess(characterName));
                    document.getElementById('cancel-attach').addEventListener('click', closeAttachPopup);
                    document.getElementById('process-search').addEventListener('input', filterProcessList);
                }
            })
            .catch(error => {
                console.error('Error fetching process list:', error);
                const popup = document.querySelector('.attach-popup');
                popup.innerHTML = `
                    <h3>Error</h3>
                    <p>An error occurred while fetching the process list.</p>
                    <button onclick="closeAttachPopup()" class="btn btn-primary">Close</button>
                `;
            });
    }

    function selectProcess(row, pid) {
        const allRows = document.querySelectorAll('#process-list-body tr');
        allRows.forEach(r => r.classList.remove('selected'));
        row.classList.add('selected');
        document.getElementById('choose-process').disabled = false;
        document.getElementById('choose-process').dataset.pid = pid;
        document.getElementById('selected-pid').textContent = pid;
    }

    function chooseProcess(characterName) {
        const pid = document.getElementById('choose-process').dataset.pid;
        if (pid) {
            // Show loading animation
            const popup = document.querySelector('.attach-popup');
            popup.innerHTML = `
                <h3>Attaching to Process</h3>
                <div class="loading-spinner"></div>
                <p>Please wait...</p>
            `;
    
            fetch(`/attach-process?characterName=${characterName}&pid=${pid}`, { method: 'POST' })
                .then(response => response.json())
                .then(data => {
                    if (data.success) {
                        // Show success message
                        popup.innerHTML = `
                            <h3>Success</h3>
                            <p>Successfully attached to process ${pid} for character ${characterName}</p>
                        `;
                        // Close popup after 2 seconds
                        setTimeout(() => {
                            closeAttachPopup();
                            fetchInitialData(); // Refresh the dashboard
                        }, 2000);
                    } else {
                        // Show error message
                        popup.innerHTML = `
                            <h3>Error</h3>
                            <p>Failed to attach to process: ${data.error}</p>
                            <button onclick="closeAttachPopup()" class="btn btn-primary">Close</button>
                        `;
                    }
                })
                .catch(error => {
                    console.error('Error attaching to process:', error);
                    // Show error message
                    popup.innerHTML = `
                        <h3>Error</h3>
                        <p>An error occurred while attaching to the process.</p>
                        <button onclick="closeAttachPopup()" class="btn btn-primary">Close</button>
                    `;
                });
        }
    }

    async function reloadConfig() {
        const btn = document.getElementById('reloadConfigBtn');
        const icon = btn.querySelector('i');
        
        // Disable button and start rotation
        btn.disabled = true;
        icon.classList.add('rotate');
        
        try {
            const response = await fetch('/api/reload-config');
            if (!response.ok) {
                throw new Error('Failed to reload config');
            }
        } catch (error) {
            console.error('Error reloading config:', error);
        } finally {
            // Re-enable button and stop rotation
            btn.disabled = false;
            icon.classList.remove('rotate');
        }
    }

    function closeAttachPopup() {
        const popup = document.querySelector('.attach-popup');
        if (popup) {
            popup.remove();
        }
    }

    function filterProcessList() {
        const searchTerm = document.getElementById('process-search').value.toLowerCase();
        const rows = document.querySelectorAll('#process-list-body tr');

        rows.forEach(row => {
            const windowTitle = row.cells[0].textContent.toLowerCase();
            const processName = row.cells[1].textContent.toLowerCase();
            if (windowTitle.includes(searchTerm) || processName.includes(searchTerm)) {
                row.style.display = '';
            } else {
                row.style.display = 'none';
            }
        });
    }

    function showCompanionJoinPopup(characterName) {
        const popup = document.createElement('div');
        popup.className = 'attach-popup'; // Reuse the attach popup styling
        popup.innerHTML = `
            <h3>Join Game as Companion</h3>
            <div class="popup-content">
                <div class="form-group">
                    <label for="game-name">Game Name:</label>
                    <input type="text" id="game-name" placeholder="Enter game name">
                </div>
                <div class="form-group">
                    <label for="game-password">Game Password:</label>
                    <input type="text" id="game-password" placeholder="Enter game password">
                </div>
                <div class="popup-buttons">
                    <button id="join-game-btn" class="btn btn-primary">Request Join</button>
                    <button id="cancel-join" class="btn">Cancel</button>
                </div>
            </div>
        `;
        
        document.body.appendChild(popup);
        
        // Add event listeners
        document.getElementById('join-game-btn').addEventListener('click', () => {
            const gameName = document.getElementById('game-name').value.trim();
            const password = document.getElementById('game-password').value.trim();
            
            if (!gameName) {
                alert('Please enter a game name');
                return;
            }
            
            requestCompanionJoin(characterName, gameName, password);
        });
        
        document.getElementById('cancel-join').addEventListener('click', closeCompanionJoinPopup);
    }

    function closeCompanionJoinPopup() {
        const popup = document.querySelector('.attach-popup');
        if (popup) {
            popup.remove();
        }
    }

    function requestCompanionJoin(supervisor, gameName, password) {
        // Show loading animation
        const popup = document.querySelector('.attach-popup');
        popup.innerHTML = `
            <h3>Requesting Game Join</h3>
            <div class="loading-spinner"></div>
            <p>Please wait...</p>
        `;

        fetch('/api/companion-join', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                supervisor: supervisor,
                gameName: gameName,
                password: password
            })
        })
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                // Show success message
                popup.innerHTML = `
                    <h3>Success</h3>
                    <p>Join request sent for game "${gameName}"</p>
                `;
                // Close popup after 2 seconds
                setTimeout(() => {
                    closeCompanionJoinPopup();
                }, 2000);
            } else {
                // Show error message
                popup.innerHTML = `
                    <h3>Error</h3>
                    <p>Failed to send join request: ${data.error || 'Unknown error'}</p>
                    <button onclick="closeCompanionJoinPopup()" class="btn btn-primary">Close</button>
                `;
            }
        })
        .catch(error => {
            console.error('Error sending join request:', error);
            // Show error message
            popup.innerHTML = `
                <h3>Error</h3>
                <p>An error occurred while sending the join request.</p>
                <button onclick="closeCompanionJoinPopup()" class="btn btn-primary">Close</button>
            `;
        });
    }

    document.addEventListener('DOMContentLoaded', function() {
        fetchInitialData();
        connectWebSocket();
        restoreExpandedState();
    });