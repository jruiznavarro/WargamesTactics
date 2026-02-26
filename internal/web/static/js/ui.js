// UI controller for the AoS battle simulator web interface.
var UI = (function() {
    var currentMode = '';
    var currentPhase = '';
    var pendingCommand = null;  // command type waiting for target/position

    function showScreen(id) {
        var screens = document.querySelectorAll('.screen');
        for (var i = 0; i < screens.length; i++) {
            screens[i].classList.remove('active');
        }
        document.getElementById(id).classList.add('active');
    }

    function updateTopBar(state) {
        if (!state) return;
        document.getElementById('round-label').textContent = 'Ronda ' + (state.BattleRound || 0);
        document.getElementById('phase-label').textContent = state.CurrentPhase || '-';

        var vp = state.VictoryPoints || {};
        var cp = state.CommandPoints || {};
        document.getElementById('vp1').textContent = vp[1] || 0;
        document.getElementById('vp2').textContent = vp[2] || 0;
        document.getElementById('cp1').textContent = cp[1] || 0;
        document.getElementById('cp2').textContent = cp[2] || 0;
    }

    function showUnitDetails(unit) {
        if (!unit) {
            document.getElementById('unit-details').innerHTML = '<p class="placeholder">Haz clic en una unidad</p>';
            return;
        }

        var isP1 = unit.OwnerID === 1;
        var html = '<div class="unit-card">';
        html += '<div class="unit-name ' + (isP1 ? 'player1' : 'player2') + '">[' + unit.ID + '] ' + unit.Name + '</div>';

        // Health bar
        var pct = unit.MaxWounds > 0 ? (unit.CurrentWounds / unit.MaxWounds * 100) : 0;
        var hpColor = pct > 60 ? '#4caf50' : pct > 30 ? '#ff9800' : '#f44336';
        html += '<div class="health-bar"><div class="fill" style="width:' + pct + '%;background:' + hpColor + '"></div></div>';

        // Stats
        html += '<div class="stat-row">';
        html += '<span>HP: <strong>' + unit.CurrentWounds + '/' + unit.MaxWounds + '</strong></span>';
        html += '<span>Modelos: <strong>' + unit.AliveModels + '/' + unit.TotalModels + '</strong></span>';
        html += '</div>';
        html += '<div class="stat-row">';
        html += '<span>Move: <strong>' + unit.MoveSpeed + '"</strong></span>';
        html += '<span>Save: <strong>' + unit.Save + '+</strong></span>';
        if (unit.WardSave > 0) html += '<span>Ward: <strong>' + unit.WardSave + '+</strong></span>';
        html += '</div>';
        html += '<div class="stat-row">';
        html += '<span>Pos: <strong>(' + unit.Position[0].toFixed(1) + ', ' + unit.Position[1].toFixed(1) + ')</strong></span>';
        html += '</div>';

        // Weapons
        if (unit.Weapons && unit.Weapons.length > 0) {
            html += '<div class="weapon-list">';
            for (var i = 0; i < unit.Weapons.length; i++) {
                var w = unit.Weapons[i];
                var rangeStr = w.Range > 0 ? w.Range + '"' : 'Melee';
                var rendStr = w.Rend > 0 ? ' R:-' + w.Rend : '';
                html += '<div class="weapon-item">' + w.Name + ' (' + rangeStr + ') A:' + w.Attacks + ' ' + w.ToHit + '+/' + w.ToWound + '+ D:' + w.Damage + rendStr + '</div>';
            }
            html += '</div>';
        }

        // Spells
        if (unit.Spells && unit.Spells.length > 0) {
            html += '<div class="weapon-list">';
            for (var i = 0; i < unit.Spells.length; i++) {
                var sp = unit.Spells[i];
                html += '<div class="weapon-item">Spell: ' + sp.Name + ' (CV:' + sp.CastingValue + ' R:' + sp.Range + '")</div>';
            }
            html += '</div>';
        }

        // Status flags
        var flags = [];
        if (unit.HasMoved) flags.push('MOVED');
        if (unit.HasRun) flags.push('RAN');
        if (unit.HasRetreated) flags.push('RETREATED');
        if (unit.HasShot) flags.push('SHOT');
        if (unit.HasFought) flags.push('FOUGHT');
        if (unit.HasCharged) flags.push('CHARGED');
        if (unit.HasPiledIn) flags.push('PILED-IN');
        if (unit.IsEngaged) flags.push('ENGAGED');
        if (unit.CanCast) flags.push('CAN CAST');
        if (unit.CanChant) flags.push('CAN CHANT');

        if (flags.length > 0) {
            html += '<div class="status-flags">';
            for (var i = 0; i < flags.length; i++) {
                var active = ['ENGAGED', 'CAN CAST', 'CAN CHANT'].indexOf(flags[i]) !== -1;
                html += '<span class="flag' + (active ? ' active' : '') + '">' + flags[i] + '</span>';
            }
            html += '</div>';
        }

        html += '</div>';
        document.getElementById('unit-details').innerHTML = html;
    }

    function showCommandPanel(phase) {
        var panel = document.getElementById('command-panel');
        var buttons = document.getElementById('command-buttons');

        if (currentMode === 'aivai') {
            panel.classList.add('hidden');
            return;
        }

        panel.classList.remove('hidden');
        buttons.innerHTML = '';

        if (!phase || !phase.AllowedCommands) return;

        var cmds = phase.AllowedCommands;
        var cmdLabels = {
            'move':     'Mover',
            'run':      'Correr',
            'retreat':  'Retirar',
            'shoot':    'Disparar',
            'charge':   'Cargar',
            'fight':    'Combatir',
            'pile_in':  'Pile In',
            'cast':     'Lanzar Hechizo',
            'chant':    'Rezar',
            'rally':    'Rally',
            'end_phase':'Terminar Fase'
        };

        for (var i = 0; i < cmds.length; i++) {
            var cmd = cmds[i];
            var btn = document.createElement('button');
            btn.className = 'cmd-btn';
            btn.textContent = cmdLabels[cmd] || cmd;
            btn.dataset.cmd = cmd;
            btn.addEventListener('click', onCommandClick);
            buttons.appendChild(btn);
        }
    }

    function onCommandClick(evt) {
        var cmd = evt.target.dataset.cmd;
        Canvas.clearClickCallback();

        // Deselect all buttons
        var btns = document.querySelectorAll('.cmd-btn');
        for (var i = 0; i < btns.length; i++) btns[i].classList.remove('active');

        if (cmd === 'end_phase') {
            App.sendCommand({ type: 'skip', data: {} });
            return;
        }

        evt.target.classList.add('active');
        pendingCommand = cmd;

        var selUnit = Canvas.getSelectedUnit();

        switch(cmd) {
        case 'move':
        case 'run':
        case 'retreat':
            if (!selUnit) {
                addLogEntry('Selecciona una unidad primero');
                return;
            }
            addLogEntry('Haz clic en el destino en el mapa');
            Canvas.setMoveMode(true);
            Canvas.setClickCallback(function(unit, boardPos) {
                Canvas.clearClickCallback();
                Canvas.setMoveMode(false);
                App.sendCommand({
                    type: cmd,
                    data: { unitId: selUnit, x: Math.round(boardPos.x * 10) / 10, y: Math.round(boardPos.y * 10) / 10 }
                });
                resetCommandButtons();
            });
            break;

        case 'shoot':
            if (!selUnit) {
                addLogEntry('Selecciona una unidad primero');
                return;
            }
            addLogEntry('Haz clic en el objetivo enemigo');
            Canvas.setTargetMode(true);
            Canvas.setClickCallback(function(unit) {
                Canvas.clearClickCallback();
                if (!unit) { addLogEntry('No hay unidad en ese punto'); resetCommandButtons(); return; }
                App.sendCommand({
                    type: 'shoot',
                    data: { shooterId: selUnit, targetId: unit.ID }
                });
                resetCommandButtons();
            });
            break;

        case 'charge':
            if (!selUnit) {
                addLogEntry('Selecciona una unidad primero');
                return;
            }
            addLogEntry('Haz clic en el objetivo de la carga');
            Canvas.setTargetMode(true);
            Canvas.setClickCallback(function(unit) {
                Canvas.clearClickCallback();
                if (!unit) { addLogEntry('No hay unidad en ese punto'); resetCommandButtons(); return; }
                App.sendCommand({
                    type: 'charge',
                    data: { chargerId: selUnit, targetId: unit.ID }
                });
                resetCommandButtons();
            });
            break;

        case 'fight':
            if (!selUnit) {
                addLogEntry('Selecciona una unidad primero');
                return;
            }
            addLogEntry('Haz clic en el objetivo del combate');
            Canvas.setTargetMode(true);
            Canvas.setClickCallback(function(unit) {
                Canvas.clearClickCallback();
                if (!unit) { addLogEntry('No hay unidad en ese punto'); resetCommandButtons(); return; }
                App.sendCommand({
                    type: 'fight',
                    data: { attackerId: selUnit, targetId: unit.ID }
                });
                resetCommandButtons();
            });
            break;

        case 'pile_in':
            if (!selUnit) {
                addLogEntry('Selecciona una unidad primero');
                return;
            }
            App.sendCommand({
                type: 'pilein',
                data: { unitId: selUnit }
            });
            resetCommandButtons();
            break;

        case 'cast':
            if (!selUnit) {
                addLogEntry('Selecciona un mago primero');
                return;
            }
            var su = Canvas.getUnit(selUnit);
            if (!su || !su.Spells || su.Spells.length === 0) {
                addLogEntry('Esta unidad no tiene hechizos');
                resetCommandButtons();
                return;
            }
            // Auto-pick first spell, then pick target
            addLogEntry('Haz clic en el objetivo del hechizo');
            Canvas.setTargetMode(true);
            Canvas.setClickCallback(function(unit) {
                Canvas.clearClickCallback();
                if (!unit) { addLogEntry('No hay unidad en ese punto'); resetCommandButtons(); return; }
                App.sendCommand({
                    type: 'cast',
                    data: { casterId: selUnit, spellIndex: 0, targetId: unit.ID }
                });
                resetCommandButtons();
            });
            break;

        case 'chant':
            if (!selUnit) {
                addLogEntry('Selecciona un sacerdote primero');
                return;
            }
            addLogEntry('Haz clic en el objetivo de la plegaria');
            Canvas.setTargetMode(true);
            Canvas.setClickCallback(function(unit) {
                Canvas.clearClickCallback();
                if (!unit) { addLogEntry('No hay unidad en ese punto'); resetCommandButtons(); return; }
                App.sendCommand({
                    type: 'chant',
                    data: { chanterId: selUnit, prayerIndex: 0, targetId: unit.ID }
                });
                resetCommandButtons();
            });
            break;

        case 'rally':
            if (!selUnit) {
                addLogEntry('Selecciona una unidad primero');
                return;
            }
            App.sendCommand({
                type: 'rally',
                data: { unitId: selUnit }
            });
            resetCommandButtons();
            break;
        }
    }

    function resetCommandButtons() {
        var btns = document.querySelectorAll('.cmd-btn');
        for (var i = 0; i < btns.length; i++) btns[i].classList.remove('active');
        pendingCommand = null;
    }

    function addLogEntry(text) {
        var log = document.getElementById('battle-log');
        var entry = document.createElement('div');
        entry.className = 'log-entry';

        // Style based on content
        if (text.indexOf('=== BATTLE ROUND') !== -1) {
            entry.className += ' round-header';
        } else if (text.indexOf('-- ') === 2 || text.indexOf('-- ') === 0) {
            entry.className += ' phase-header';
        }

        entry.textContent = text;
        log.appendChild(entry);
        log.scrollTop = log.scrollHeight;
    }

    function addLogEntries(entries) {
        for (var i = 0; i < entries.length; i++) {
            addLogEntry(entries[i]);
        }
    }

    function showGameOver(data) {
        var overlay = document.getElementById('game-over');
        overlay.classList.remove('hidden');

        var title = document.getElementById('game-over-title');
        var message = document.getElementById('game-over-message');
        var scores = document.getElementById('game-over-scores');

        if (data.winner >= 0 && data.winnerName) {
            title.textContent = 'Victoria!';
            message.textContent = data.winnerName + ' (Jugador ' + data.winner + ') gana la batalla!';
        } else {
            title.textContent = 'Empate';
            message.textContent = 'La batalla termina sin un ganador claro.';
        }

        var vp = data.victoryPoints || {};
        scores.innerHTML = 'VP - P1: ' + (vp[1] || 0) + ' | P2: ' + (vp[2] || 0);
    }

    function setMode(mode) {
        currentMode = mode;
    }

    function setFactionNames(f1, f2) {
        document.getElementById('faction1-name').textContent = f1;
        document.getElementById('faction2-name').textContent = f2;
    }

    return {
        showScreen: showScreen,
        updateTopBar: updateTopBar,
        showUnitDetails: showUnitDetails,
        showCommandPanel: showCommandPanel,
        addLogEntry: addLogEntry,
        addLogEntries: addLogEntries,
        showGameOver: showGameOver,
        setMode: setMode,
        setFactionNames: setFactionNames,
        resetCommandButtons: resetCommandButtons
    };
})();
