// Main application entry point for the AoS battle simulator web interface.
var App = (function() {
    var connected = false;

    function init() {
        Canvas.init();

        // Load factions
        fetch('/api/factions')
            .then(function(r) { return r.json(); })
            .then(function(factions) {
                var sel1 = document.getElementById('faction1');
                var sel2 = document.getElementById('faction2');
                for (var i = 0; i < factions.length; i++) {
                    var f = factions[i];
                    var opt1 = document.createElement('option');
                    opt1.value = f.id;
                    opt1.textContent = f.name;
                    sel1.appendChild(opt1);

                    var opt2 = document.createElement('option');
                    opt2.value = f.id;
                    opt2.textContent = f.name;
                    sel2.appendChild(opt2);
                }
                // Default second faction to a different one if possible
                if (factions.length > 1) sel2.selectedIndex = 1;
            })
            .catch(function(err) {
                console.error('Failed to load factions:', err);
            });

        // Wire up events
        document.getElementById('btn-start').addEventListener('click', startGame);
        document.getElementById('btn-new-game').addEventListener('click', function() {
            location.reload();
        });

        // Register WebSocket handlers
        WS.on('game_started', onGameStarted);
        WS.on('game_state', onGameState);
        WS.on('battle_log', onBattleLog);
        WS.on('prompt', onPrompt);
        WS.on('game_over', onGameOver);
        WS.on('error', onError);
        WS.on('pong', function() {});
    }

    function startGame() {
        var mode = document.getElementById('mode').value;
        var faction1 = document.getElementById('faction1').value;
        var faction2 = document.getElementById('faction2').value;
        var seed = parseInt(document.getElementById('seed').value) || 0;
        var rounds = parseInt(document.getElementById('rounds').value) || 5;

        if (!faction1 || !faction2) {
            alert('Selecciona las dos facciones');
            return;
        }

        document.getElementById('btn-start').disabled = true;
        document.getElementById('btn-start').textContent = 'Conectando...';

        UI.setMode(mode);

        WS.connect(function() {
            connected = true;
            WS.send('start_game', {
                mode: mode,
                faction1: faction1,
                faction2: faction2,
                seed: seed,
                rounds: rounds
            });
        });
    }

    function onGameStarted(data) {
        UI.showScreen('game-screen');
        UI.setFactionNames(data.faction1, data.faction2);
        UI.addLogEntry('Batalla iniciada: ' + data.faction1 + ' vs ' + data.faction2);
        UI.addLogEntry('Battleplan: ' + data.battleplan + ' | Seed: ' + data.seed);
    }

    function onGameState(data) {
        Canvas.update(data);
        UI.updateTopBar(data);
    }

    function onBattleLog(entries) {
        UI.addLogEntries(entries);
    }

    function onPrompt(data) {
        // The game wants a command from the player
        Canvas.update(data.view || data.View);
        UI.updateTopBar(data.view || data.View);

        var phase = data.phase || data.Phase;
        UI.showCommandPanel(phase);
        UI.addLogEntry('--- Tu turno: ' + (phase.Type || '') + ' ---');
    }

    function onGameOver(data) {
        UI.showGameOver(data);
    }

    function onError(data) {
        var msg = data.message || data;
        UI.addLogEntry('ERROR: ' + msg);
        console.error('Server error:', msg);
    }

    function sendCommand(wsCmd) {
        WS.send('command', wsCmd);
    }

    // Initialize on DOM ready
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', init);
    } else {
        init();
    }

    return {
        sendCommand: sendCommand
    };
})();
