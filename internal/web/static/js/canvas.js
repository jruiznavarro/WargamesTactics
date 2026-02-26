// Canvas 2D battlefield renderer for the AoS battle simulator.
var Canvas = (function() {
    var canvas, ctx;
    var boardWidth = 60, boardHeight = 44;
    var canvasW = 960, canvasH = 540;
    var units = {};       // unitID -> unit data
    var terrain = [];
    var objectives = [];
    var territories = [];
    var selectedUnitId = null;
    var hoveredUnitId = null;
    var targetMode = false;  // true when waiting for target click
    var moveMode = false;    // true when waiting for move destination click

    // Faction color palette
    var P1_COLOR = '#4fc3f7';
    var P1_FILL = 'rgba(79, 195, 247, 0.7)';
    var P2_COLOR = '#ef5350';
    var P2_FILL = 'rgba(239, 83, 80, 0.7)';
    var TERRAIN_COLORS = {
        'Obscuring': { fill: 'rgba(34, 139, 34, 0.4)', stroke: '#228b22' },
        'Obstacle':  { fill: 'rgba(139, 119, 101, 0.4)', stroke: '#8b7765' },
        'Area':      { fill: 'rgba(210, 180, 140, 0.3)', stroke: '#d2b48c' },
        'Impassable':{ fill: 'rgba(139, 0, 0, 0.4)', stroke: '#8b0000' }
    };

    // Keyword-based shapes
    var UNIT_SHAPES = {
        // default: circle
        'Hero':     'diamond',
        'Monster':  'hexagon',
        'Cavalry':  'triangle',
        'War Machine': 'square'
    };

    function init() {
        canvas = document.getElementById('battlefield');
        ctx = canvas.getContext('2d');
        canvas.width = canvasW;
        canvas.height = canvasH;

        canvas.addEventListener('mousemove', onMouseMove);
        canvas.addEventListener('click', onClick);
    }

    function setBoardSize(w, h) {
        boardWidth = w;
        boardHeight = h;
    }

    function toCanvas(x, y) {
        var margin = 20;
        var usableW = canvasW - 2 * margin;
        var usableH = canvasH - 2 * margin;
        return {
            x: margin + (x / boardWidth) * usableW,
            y: margin + (y / boardHeight) * usableH
        };
    }

    function toBoard(cx, cy) {
        var margin = 20;
        var usableW = canvasW - 2 * margin;
        var usableH = canvasH - 2 * margin;
        return {
            x: ((cx - margin) / usableW) * boardWidth,
            y: ((cy - margin) / usableH) * boardHeight
        };
    }

    function update(gameState) {
        if (!gameState) return;

        if (gameState.BoardWidth) boardWidth = gameState.BoardWidth;
        if (gameState.BoardHeight) boardHeight = gameState.BoardHeight;

        // Flatten units
        units = {};
        if (gameState.Units) {
            for (var ownerID in gameState.Units) {
                var list = gameState.Units[ownerID];
                for (var i = 0; i < list.length; i++) {
                    var u = list[i];
                    u.OwnerID = parseInt(ownerID);
                    units[u.ID] = u;
                }
            }
        }

        terrain = gameState.Terrain || [];
        objectives = gameState.Objectives || [];
        territories = gameState.Territories || [];

        render();
    }

    function render() {
        ctx.clearRect(0, 0, canvasW, canvasH);
        drawBackground();
        drawTerritories();
        drawGrid();
        drawTerrain();
        drawObjectives();
        drawUnits();
        drawMoveRange();
    }

    function drawBackground() {
        // Gradient background like a battlefield
        var grad = ctx.createLinearGradient(0, 0, 0, canvasH);
        grad.addColorStop(0, '#1a2a1a');
        grad.addColorStop(0.5, '#1e2e1e');
        grad.addColorStop(1, '#1a2a1a');
        ctx.fillStyle = grad;
        ctx.fillRect(0, 0, canvasW, canvasH);
    }

    function drawTerritories() {
        for (var i = 0; i < territories.length; i++) {
            var t = territories[i];
            if (!t.MinPos || !t.MaxPos) continue;
            var p1 = toCanvas(t.MinPos[0], t.MinPos[1]);
            var p2 = toCanvas(t.MaxPos[0], t.MaxPos[1]);
            ctx.fillStyle = i === 0 ? 'rgba(79, 195, 247, 0.05)' : 'rgba(239, 83, 80, 0.05)';
            ctx.strokeStyle = i === 0 ? 'rgba(79, 195, 247, 0.15)' : 'rgba(239, 83, 80, 0.15)';
            ctx.lineWidth = 1;
            ctx.setLineDash([4, 4]);
            ctx.fillRect(p1.x, p1.y, p2.x - p1.x, p2.y - p1.y);
            ctx.strokeRect(p1.x, p1.y, p2.x - p1.x, p2.y - p1.y);
            ctx.setLineDash([]);
        }
    }

    function drawGrid() {
        ctx.strokeStyle = 'rgba(255, 255, 255, 0.06)';
        ctx.lineWidth = 0.5;
        // Draw grid lines every 6 inches
        var step = 6;
        for (var x = 0; x <= boardWidth; x += step) {
            var p = toCanvas(x, 0);
            var p2 = toCanvas(x, boardHeight);
            ctx.beginPath();
            ctx.moveTo(p.x, p.y);
            ctx.lineTo(p2.x, p2.y);
            ctx.stroke();
        }
        for (var y = 0; y <= boardHeight; y += step) {
            var p = toCanvas(0, y);
            var p2 = toCanvas(boardWidth, y);
            ctx.beginPath();
            ctx.moveTo(p.x, p.y);
            ctx.lineTo(p2.x, p2.y);
            ctx.stroke();
        }

        // Board border
        var tl = toCanvas(0, 0);
        var br = toCanvas(boardWidth, boardHeight);
        ctx.strokeStyle = '#0f3460';
        ctx.lineWidth = 2;
        ctx.strokeRect(tl.x, tl.y, br.x - tl.x, br.y - tl.y);
    }

    function drawTerrain() {
        for (var i = 0; i < terrain.length; i++) {
            var t = terrain[i];
            var p1 = toCanvas(t.Pos[0], t.Pos[1]);
            var p2 = toCanvas(t.Pos[0] + t.Width, t.Pos[1] + t.Height);
            var colors = TERRAIN_COLORS[t.Type] || { fill: 'rgba(100,100,100,0.3)', stroke: '#666' };

            ctx.fillStyle = colors.fill;
            ctx.strokeStyle = colors.stroke;
            ctx.lineWidth = 1.5;

            // Rounded rectangle for terrain
            var rx = p1.x, ry = p1.y;
            var rw = p2.x - p1.x, rh = p2.y - p1.y;
            var r = 4;
            ctx.beginPath();
            ctx.moveTo(rx + r, ry);
            ctx.lineTo(rx + rw - r, ry);
            ctx.quadraticCurveTo(rx + rw, ry, rx + rw, ry + r);
            ctx.lineTo(rx + rw, ry + rh - r);
            ctx.quadraticCurveTo(rx + rw, ry + rh, rx + rw - r, ry + rh);
            ctx.lineTo(rx + r, ry + rh);
            ctx.quadraticCurveTo(rx, ry + rh, rx, ry + rh - r);
            ctx.lineTo(rx, ry + r);
            ctx.quadraticCurveTo(rx, ry, rx + r, ry);
            ctx.closePath();
            ctx.fill();
            ctx.stroke();

            // Terrain label
            ctx.fillStyle = colors.stroke;
            ctx.font = '10px sans-serif';
            ctx.textAlign = 'center';
            ctx.fillText(t.Name, (p1.x + p2.x) / 2, (p1.y + p2.y) / 2 + 3);
        }
    }

    function drawObjectives() {
        for (var i = 0; i < objectives.length; i++) {
            var o = objectives[i];
            var p = toCanvas(o.Position[0], o.Position[1]);
            var r = Math.max(8, (o.Radius / boardWidth) * (canvasW - 40));

            // Objective circle
            ctx.beginPath();
            ctx.arc(p.x, p.y, r, 0, Math.PI * 2);
            if (o.ControlledBy === 1) {
                ctx.fillStyle = 'rgba(79, 195, 247, 0.2)';
                ctx.strokeStyle = P1_COLOR;
            } else if (o.ControlledBy === 2) {
                ctx.fillStyle = 'rgba(239, 83, 80, 0.2)';
                ctx.strokeStyle = P2_COLOR;
            } else {
                ctx.fillStyle = 'rgba(255, 215, 0, 0.15)';
                ctx.strokeStyle = '#ffd700';
            }
            ctx.lineWidth = 2;
            ctx.fill();
            ctx.stroke();

            // Star icon in center
            drawStar(p.x, p.y, 4, ctx.strokeStyle);
        }
    }

    function drawStar(cx, cy, size, color) {
        ctx.fillStyle = color;
        ctx.beginPath();
        for (var i = 0; i < 5; i++) {
            var angle = (i * 4 * Math.PI / 5) - Math.PI / 2;
            var x = cx + Math.cos(angle) * size;
            var y = cy + Math.sin(angle) * size;
            if (i === 0) ctx.moveTo(x, y);
            else ctx.lineTo(x, y);
        }
        ctx.closePath();
        ctx.fill();
    }

    function drawUnits() {
        for (var id in units) {
            var u = units[id];
            var p = toCanvas(u.Position[0], u.Position[1]);
            var isP1 = u.OwnerID === 1;
            var color = isP1 ? P1_COLOR : P2_COLOR;
            var fillColor = isP1 ? P1_FILL : P2_FILL;
            var size = 12;

            // Larger for monsters/heroes
            if (hasKeywordLike(u, 'Monster')) size = 18;
            else if (hasKeywordLike(u, 'Hero')) size = 15;
            else if (hasKeywordLike(u, 'Cavalry')) size = 14;

            var isSelected = (selectedUnitId === u.ID);
            var isHovered = (hoveredUnitId === u.ID);

            // Selection glow
            if (isSelected) {
                ctx.beginPath();
                ctx.arc(p.x, p.y, size + 6, 0, Math.PI * 2);
                ctx.fillStyle = isP1 ? 'rgba(79, 195, 247, 0.25)' : 'rgba(239, 83, 80, 0.25)';
                ctx.fill();
                ctx.strokeStyle = color;
                ctx.lineWidth = 2;
                ctx.setLineDash([3, 3]);
                ctx.stroke();
                ctx.setLineDash([]);
            }

            // Hover glow
            if (isHovered && !isSelected) {
                ctx.beginPath();
                ctx.arc(p.x, p.y, size + 4, 0, Math.PI * 2);
                ctx.fillStyle = 'rgba(255, 255, 255, 0.1)';
                ctx.fill();
            }

            // Unit shape
            var shape = getUnitShape(u);
            drawShape(p.x, p.y, size, shape, fillColor, color, isSelected ? 2.5 : 1.5);

            // Unit ID label
            ctx.fillStyle = '#fff';
            ctx.font = 'bold 10px sans-serif';
            ctx.textAlign = 'center';
            ctx.textBaseline = 'middle';
            ctx.fillText(u.ID, p.x, p.y);

            // Health bar below unit
            if (u.MaxWounds > 0) {
                var barW = size * 1.6;
                var barH = 3;
                var barX = p.x - barW / 2;
                var barY = p.y + size + 4;
                var pct = u.CurrentWounds / u.MaxWounds;

                ctx.fillStyle = '#333';
                ctx.fillRect(barX, barY, barW, barH);

                var hpColor = pct > 0.6 ? '#4caf50' : pct > 0.3 ? '#ff9800' : '#f44336';
                ctx.fillStyle = hpColor;
                ctx.fillRect(barX, barY, barW * pct, barH);
            }

            // Model count
            if (u.TotalModels > 1) {
                ctx.fillStyle = '#aaa';
                ctx.font = '8px sans-serif';
                ctx.textAlign = 'center';
                ctx.fillText(u.AliveModels + '/' + u.TotalModels, p.x, p.y + size + 13);
            }

            // Unit name
            ctx.fillStyle = color;
            ctx.font = '9px sans-serif';
            ctx.textAlign = 'center';
            ctx.fillText(u.Name, p.x, p.y - size - 5);

            // Engagement indicator
            if (u.IsEngaged) {
                ctx.beginPath();
                ctx.arc(p.x + size + 2, p.y - size + 2, 3, 0, Math.PI * 2);
                ctx.fillStyle = '#ff5722';
                ctx.fill();
            }
        }
    }

    function drawMoveRange() {
        if (!moveMode || !selectedUnitId || !units[selectedUnitId]) return;
        var u = units[selectedUnitId];
        var p = toCanvas(u.Position[0], u.Position[1]);
        var moveInches = u.MoveSpeed || 5;
        // Convert inches to canvas pixels
        var margin = 20;
        var usableW = canvasW - 2 * margin;
        var movePixels = (moveInches / boardWidth) * usableW;

        ctx.beginPath();
        ctx.arc(p.x, p.y, movePixels, 0, Math.PI * 2);
        ctx.strokeStyle = 'rgba(255, 255, 255, 0.3)';
        ctx.lineWidth = 1;
        ctx.setLineDash([5, 5]);
        ctx.stroke();
        ctx.setLineDash([]);
    }

    function hasKeywordLike(unit, keyword) {
        // Check via Weapons range or unit name heuristics
        // In the WebSocket data we don't always have keywords, so also check name
        if (unit.Name && unit.Name.toLowerCase().indexOf(keyword.toLowerCase()) !== -1) return true;
        return false;
    }

    function getUnitShape(u) {
        // Simple heuristic based on unit properties
        if (u.MoveSpeed >= 10) return 'triangle';  // Cavalry
        if (u.MaxWounds > 20) return 'hexagon';     // Monster
        if (u.TotalModels === 1 && u.MaxWounds > 5) return 'diamond'; // Hero
        return 'circle';
    }

    function drawShape(x, y, size, shape, fill, stroke, lineWidth) {
        ctx.fillStyle = fill;
        ctx.strokeStyle = stroke;
        ctx.lineWidth = lineWidth;
        ctx.beginPath();

        switch(shape) {
        case 'diamond':
            ctx.moveTo(x, y - size);
            ctx.lineTo(x + size, y);
            ctx.lineTo(x, y + size);
            ctx.lineTo(x - size, y);
            break;
        case 'triangle':
            ctx.moveTo(x, y - size);
            ctx.lineTo(x + size, y + size * 0.7);
            ctx.lineTo(x - size, y + size * 0.7);
            break;
        case 'hexagon':
            for (var i = 0; i < 6; i++) {
                var angle = (Math.PI / 3) * i - Math.PI / 6;
                var px = x + Math.cos(angle) * size;
                var py = y + Math.sin(angle) * size;
                if (i === 0) ctx.moveTo(px, py);
                else ctx.lineTo(px, py);
            }
            break;
        case 'square':
            ctx.rect(x - size * 0.8, y - size * 0.8, size * 1.6, size * 1.6);
            break;
        default: // circle
            ctx.arc(x, y, size, 0, Math.PI * 2);
        }
        ctx.closePath();
        ctx.fill();
        ctx.stroke();
    }

    function findUnitAt(canvasX, canvasY) {
        var closest = null;
        var closestDist = Infinity;
        for (var id in units) {
            var u = units[id];
            var p = toCanvas(u.Position[0], u.Position[1]);
            var dx = canvasX - p.x;
            var dy = canvasY - p.y;
            var dist = Math.sqrt(dx * dx + dy * dy);
            if (dist < 20 && dist < closestDist) {
                closest = u;
                closestDist = dist;
            }
        }
        return closest;
    }

    function onMouseMove(evt) {
        var rect = canvas.getBoundingClientRect();
        var mx = (evt.clientX - rect.left) * (canvasW / rect.width);
        var my = (evt.clientY - rect.top) * (canvasH / rect.height);

        var u = findUnitAt(mx, my);
        var newHovered = u ? u.ID : null;
        if (newHovered !== hoveredUnitId) {
            hoveredUnitId = newHovered;
            render();
        }

        // Update tooltip
        var tooltip = document.getElementById('unit-tooltip');
        if (u) {
            tooltip.classList.remove('hidden');
            tooltip.style.left = (evt.clientX - canvas.getBoundingClientRect().left + 15) + 'px';
            tooltip.style.top = (evt.clientY - canvas.getBoundingClientRect().top - 10) + 'px';
            tooltip.innerHTML = '<b>[' + u.ID + '] ' + u.Name + '</b><br>' +
                'HP: ' + u.CurrentWounds + '/' + u.MaxWounds +
                ' | Models: ' + u.AliveModels + '/' + u.TotalModels +
                '<br>Move: ' + u.MoveSpeed + '" | Save: ' + u.Save + '+' +
                (u.IsEngaged ? '<br><span style="color:#ff5722">ENGAGED</span>' : '');
        } else {
            tooltip.classList.add('hidden');
        }

        // Update cursor
        if (moveMode) {
            canvas.style.cursor = 'crosshair';
        } else if (targetMode) {
            canvas.style.cursor = u ? 'pointer' : 'crosshair';
        } else {
            canvas.style.cursor = u ? 'pointer' : 'default';
        }
    }

    var onClickCallback = null;

    function onClick(evt) {
        var rect = canvas.getBoundingClientRect();
        var mx = (evt.clientX - rect.left) * (canvasW / rect.width);
        var my = (evt.clientY - rect.top) * (canvasH / rect.height);

        if (onClickCallback) {
            var u = findUnitAt(mx, my);
            var boardPos = toBoard(mx, my);
            onClickCallback(u, boardPos);
            return;
        }

        var u = findUnitAt(mx, my);
        if (u) {
            selectedUnitId = u.ID;
            render();
            if (typeof UI !== 'undefined') UI.showUnitDetails(u);
        }
    }

    function setClickCallback(cb) {
        onClickCallback = cb;
    }

    function clearClickCallback() {
        onClickCallback = null;
        targetMode = false;
        moveMode = false;
    }

    function setTargetMode(on) {
        targetMode = on;
    }

    function setMoveMode(on) {
        moveMode = on;
        render();
    }

    function setSelectedUnit(id) {
        selectedUnitId = id;
        render();
    }

    function getSelectedUnit() {
        return selectedUnitId;
    }

    function getUnit(id) {
        return units[id];
    }

    function getAllUnits() {
        return units;
    }

    return {
        init: init,
        update: update,
        render: render,
        setBoardSize: setBoardSize,
        setClickCallback: setClickCallback,
        clearClickCallback: clearClickCallback,
        setTargetMode: setTargetMode,
        setMoveMode: setMoveMode,
        setSelectedUnit: setSelectedUnit,
        getSelectedUnit: getSelectedUnit,
        getUnit: getUnit,
        getAllUnits: getAllUnits
    };
})();
