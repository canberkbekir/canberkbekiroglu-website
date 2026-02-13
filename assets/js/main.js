// ==================== GLOBAL DATA ====================
let portfolioItems = [];
let experience = [];
let education = [];
let gamesShowcase = { highlighted: [], inDevelopment: [] };
let themes = {};
let siteConfig = null;
let socialLinks = []; 
let footerConfig = null;
let instagram = null;
let projectDetails = {};
let navigation = [];

let currentSection = 'portfolio';
let currentFilter = 'highlighted';
let themeColor = '#bc13fe'; // Current Hex String
let dataLoaded = false;

// Animation State for Color Transition
let currentThemeRgb = { r: 188, g: 19, b: 254 }; // Initial RGB
let targetThemeRgb = { r: 188, g: 19, b: 254 };  // Target RGB
let isTransitioningColor = false;

// Base path for data files
const DATA_PATH = './assets/data/';

// ==================== DATA LOADING ====================
async function fetchJSON(filename, params = null) {
    try {
        let url = DATA_PATH + filename;
        if (params) {
            url += '?' + params;
        }
        
        const response = await fetch(url);
        if (!response.ok) {
            throw new Error('Failed to load ' + filename);
        }
        return await response.json();
    } catch (error) {
        console.error('Error loading ' + filename + ':', error);
        return null;
    }
}

async function loadData() {
    try {
        // 1. Fetch site-config.json FIRST with a timestamp to bust cache.
        const siteConfigData = await fetchJSON('site-config.json', 't=' + new Date().getTime());
        
        if (!siteConfigData) {
            throw new Error('Critical: Could not load site-config.json');
        }
        
        siteConfig = siteConfigData;
        
        // 2. Extract the data version to use for all other files
        const dataVersion = siteConfig.versions && siteConfig.versions.data ? siteConfig.versions.data : '1.0';
        const vParam = 'v=' + dataVersion;

        console.log('Loading data with version:', dataVersion);

        // 3. Fetch all other data files using the version from site-config
        const [
            portfolioItemsData,
            experienceData,
            educationData,
            gamesShowcaseData,
            socialLinksData,
            footerData,
            instagramData,
            projectDetailsData,
            navigationData
        ] = await Promise.all([
            fetchJSON('portfolio-items.json', vParam),
            fetchJSON('experience.json', vParam),
            fetchJSON('education.json', vParam),
            fetchJSON('games-showcase.json', vParam),
            fetchJSON('social-links.json', vParam),
            fetchJSON('footer.json', vParam),
            fetchJSON('instagram.json', vParam),
            fetchJSON('project-details.json', vParam),
            fetchJSON('navigation.json', vParam)
        ]);

        portfolioItems = (portfolioItemsData || []).filter(item => item.visible !== false);
        experience = experienceData || [];
        education = educationData || [];
        gamesShowcase = gamesShowcaseData || { highlighted: [], inDevelopment: [] };
        socialLinks = socialLinksData || []; 
        footerConfig = footerData || null;
        instagram = instagramData || null;
        projectDetails = projectDetailsData || {};
        navigation = (navigationData || []).filter(item => item.visible !== false).sort((a, b) => a.order - b.order);

        // Theme colors configuration
        const defaultThemes = {
            about: { hex: '#00f3ff' },
            games: { hex: '#ff0055' },
            portfolio: { hex: '#bc13fe' },
            experience: { hex: '#ffd700' },
            education: { hex: '#0051ff' },
            more: { hex: '#ff6600' }
        };
        themes = siteConfig?.themes || defaultThemes;

        const defaultNav = navigation.find(n => n.default) || navigation[0];
        if (defaultNav) {
            currentSection = defaultNav.section;
            // Initialize colors immediately without transition for the first load
            const startHex = themes[currentSection]?.hex || '#bc13fe';
            themeColor = startHex;
            const startRgb = hexToRgb(startHex);
            if(startRgb) {
                currentThemeRgb = startRgb;
                targetThemeRgb = startRgb;
            }
        }

        dataLoaded = true;
        console.log('All Data Loaded Successfully');
        return true;
    } catch (error) {
        console.error('Error loading data:', error);
        dataLoaded = false;
        return false;
    }
}

// ==================== COLOR UTILITIES & ANIMATION ====================

function hexToRgb(hex) {
    var shorthandRegex = /^#?([a-f\d])([a-f\d])([a-f\d])$/i;
    hex = hex.replace(shorthandRegex, function(m, r, g, b) {
        return r + r + g + g + b + b;
    });
    var result = /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(hex);
    return result ? {
        r: parseInt(result[1], 16),
        g: parseInt(result[2], 16),
        b: parseInt(result[3], 16)
    } : null;
}

function rgbToHex(r, g, b) {
    return "#" + ((1 << 24) + (r << 16) + (g << 8) + b).toString(16).slice(1);
}

function lerp(start, end, t) {
    return start * (1 - t) + end * t;
}

function processColorTransition() {
    if (!isTransitioningColor) return;

    // Transition speed (0.05 is smooth and slow)
    const speed = 0.05; 
    
    // Check if close enough to snap to target
    if (Math.abs(targetThemeRgb.r - currentThemeRgb.r) < 0.5 &&
        Math.abs(targetThemeRgb.g - currentThemeRgb.g) < 0.5 &&
        Math.abs(targetThemeRgb.b - currentThemeRgb.b) < 0.5) {
        
        currentThemeRgb = { ...targetThemeRgb };
        isTransitioningColor = false;
    } else {
        currentThemeRgb.r = lerp(currentThemeRgb.r, targetThemeRgb.r, speed);
        currentThemeRgb.g = lerp(currentThemeRgb.g, targetThemeRgb.g, speed);
        currentThemeRgb.b = lerp(currentThemeRgb.b, targetThemeRgb.b, speed);
        requestAnimationFrame(processColorTransition);
    }

    // Convert current RGB to Hex
    const newHex = rgbToHex(Math.round(currentThemeRgb.r), Math.round(currentThemeRgb.g), Math.round(currentThemeRgb.b));
    
    // Update global variable
    themeColor = newHex; 
    
    // Update CSS Variable
    document.documentElement.style.setProperty('--theme-color', newHex);
    
    // Update elements that rely on JS values (gradients, active states)
    updateDynamicColors();
    updateFilterStyles();
}

// ==================== BACKGROUND ANIMATION (RESTORED) ====================
function initBackground() {
    const canvas = document.getElementById('bg-canvas');
    if (!canvas) return;

    const ctx = canvas.getContext('2d');
    let time = 0;

    const mouse = { x: null, y: null, radius: 250 };

    document.addEventListener('mousemove', function(e) {
        mouse.x = e.clientX;
        mouse.y = e.clientY;
    });

    document.addEventListener('mouseleave', function() {
        mouse.x = null;
        mouse.y = null;
    });

    const dotSpacing = 12;
    let gridDots = [];

    const shapes = [
        { type: 'controller', x: 0.1 + Math.random() * 0.25, y: 0.15 + Math.random() * 0.3, scale: 45 + Math.random() * 20, rx: Math.random() * Math.PI, ry: Math.random() * Math.PI, speedX: 0.001 + Math.random() * 0.001, speedY: 0.0012 + Math.random() * 0.001, vx: (Math.random() - 0.5) * 0.00003, vy: (Math.random() - 0.5) * 0.00002, baseX: 0, baseY: 0, driftAngle: Math.random() * Math.PI * 2, driftSpeed: 0.001 + Math.random() * 0.001, driftRadius: 0.005 + Math.random() * 0.005 },
        { type: 'codetag', x: 0.65 + Math.random() * 0.25, y: 0.2 + Math.random() * 0.35, scale: 50 + Math.random() * 20, rx: Math.random() * Math.PI, ry: Math.random() * Math.PI, speedX: -0.0008 - Math.random() * 0.001, speedY: 0.001 + Math.random() * 0.001, vx: (Math.random() - 0.5) * 0.00003, vy: (Math.random() - 0.5) * 0.00002, baseX: 0, baseY: 0, driftAngle: Math.random() * Math.PI * 2, driftSpeed: 0.0008 + Math.random() * 0.001, driftRadius: 0.004 + Math.random() * 0.006 },
        { type: 'cube', x: 0.7 + Math.random() * 0.2, y: 0.6 + Math.random() * 0.25, scale: 35 + Math.random() * 20, rx: Math.random() * Math.PI, ry: Math.random() * Math.PI, speedX: 0.0012 + Math.random() * 0.001, speedY: -0.0008 - Math.random() * 0.001, vx: (Math.random() - 0.5) * 0.00004, vy: (Math.random() - 0.5) * 0.00003, baseX: 0, baseY: 0, driftAngle: Math.random() * Math.PI * 2, driftSpeed: 0.0012 + Math.random() * 0.0008, driftRadius: 0.003 + Math.random() * 0.004 },
        { type: 'terminal', x: 0.1 + Math.random() * 0.2, y: 0.55 + Math.random() * 0.3, scale: 40 + Math.random() * 20, rx: Math.random() * Math.PI, ry: Math.random() * Math.PI, speedX: -0.0007 - Math.random() * 0.0008, speedY: 0.001 + Math.random() * 0.001, vx: (Math.random() - 0.5) * 0.00003, vy: (Math.random() - 0.5) * 0.00002, baseX: 0, baseY: 0, driftAngle: Math.random() * Math.PI * 2, driftSpeed: 0.0007 + Math.random() * 0.0008, driftRadius: 0.005 + Math.random() * 0.005 },
        { type: 'potion', x: 0.4 + Math.random() * 0.2, y: 0.7 + Math.random() * 0.2, scale: 30 + Math.random() * 15, rx: Math.random() * Math.PI, ry: Math.random() * Math.PI, speedX: 0.0015 + Math.random() * 0.001, speedY: -0.0005 - Math.random() * 0.001, vx: (Math.random() - 0.5) * 0.00004, vy: (Math.random() - 0.5) * 0.00003, baseX: 0, baseY: 0, driftAngle: Math.random() * Math.PI * 2, driftSpeed: 0.0009 + Math.random() * 0.001, driftRadius: 0.004 + Math.random() * 0.005 }
    ];

    shapes.forEach(function(shape) {
        shape.baseX = shape.x;
        shape.baseY = shape.y;
    });

    const controllerVerts = [
        {x:-2, y:-0.5, z:0.3}, {x:-1.5, y:-0.8, z:0.3}, {x:1.5, y:-0.8, z:0.3}, {x:2, y:-0.5, z:0.3},
        {x:2, y:0.3, z:0.3}, {x:1.5, y:0.6, z:0.3}, {x:-1.5, y:0.6, z:0.3}, {x:-2, y:0.3, z:0.3},
        {x:-2, y:-0.5, z:-0.3}, {x:-1.5, y:-0.8, z:-0.3}, {x:1.5, y:-0.8, z:-0.3}, {x:2, y:-0.5, z:-0.3},
        {x:2, y:0.3, z:-0.3}, {x:1.5, y:0.6, z:-0.3}, {x:-1.5, y:0.6, z:-0.3}, {x:-2, y:0.3, z:-0.3},
        {x:-1.2, y:-0.1, z:0.4}, {x:-0.8, y:-0.1, z:0.4}, {x:-0.8, y:0.3, z:0.4}, {x:-1.2, y:0.3, z:0.4},
        {x:1, y:0, z:0.4}, {x:1.3, y:-0.2, z:0.4}, {x:1.3, y:0.2, z:0.4}, {x:1, y:0.4, z:0.4},
        {x:-1, y:-0.3, z:0.35}, {x:-1, y:0.1, z:0.35}, {x:-1.2, y:-0.1, z:0.35}, {x:-0.8, y:-0.1, z:0.35}
    ];
    const controllerEdges = [[0,1],[1,2],[2,3],[3,4],[4,5],[5,6],[6,7],[7,0],[8,9],[9,10],[10,11],[11,12],[12,13],[13,14],[14,15],[15,8],[0,8],[1,9],[2,10],[3,11],[4,12],[5,13],[6,14],[7,15],[16,17],[17,18],[18,19],[19,16],[20,21],[21,22],[22,23],[23,20],[24,25],[26,27]];

    const codetagVerts = [
        {x:-1.5, y:0, z:0.2}, {x:-0.5, y:1, z:0.2}, {x:-0.5, y:0.6, z:0.2}, {x:-1.1, y:0, z:0.2}, {x:-0.5, y:-0.6, z:0.2}, {x:-0.5, y:-1, z:0.2},
        {x:-1.5, y:0, z:-0.2}, {x:-0.5, y:1, z:-0.2}, {x:-0.5, y:0.6, z:-0.2}, {x:-1.1, y:0, z:-0.2}, {x:-0.5, y:-0.6, z:-0.2}, {x:-0.5, y:-1, z:-0.2},
        {x:-0.15, y:1.2, z:0.2}, {x:0.15, y:1.2, z:0.2}, {x:0.15, y:-1.2, z:0.2}, {x:-0.15, y:-1.2, z:0.2},
        {x:-0.15, y:1.2, z:-0.2}, {x:0.15, y:1.2, z:-0.2}, {x:0.15, y:-1.2, z:-0.2}, {x:-0.15, y:-1.2, z:-0.2},
        {x:1.5, y:0, z:0.2}, {x:0.5, y:1, z:0.2}, {x:0.5, y:0.6, z:0.2}, {x:1.1, y:0, z:0.2}, {x:0.5, y:-0.6, z:0.2}, {x:0.5, y:-1, z:0.2},
        {x:1.5, y:0, z:-0.2}, {x:0.5, y:1, z:-0.2}, {x:0.5, y:0.6, z:-0.2}, {x:1.1, y:0, z:-0.2}, {x:0.5, y:-0.6, z:-0.2}, {x:0.5, y:-1, z:-0.2}
    ];
    const codetagEdges = [[0,1],[1,2],[2,3],[3,4],[4,5],[5,0],[6,7],[7,8],[8,9],[9,10],[10,11],[11,6],[0,6],[1,7],[5,11],[12,13],[13,14],[14,15],[15,12],[16,17],[17,18],[18,19],[19,16],[12,16],[13,17],[14,18],[15,19],[20,21],[21,22],[22,23],[23,24],[24,25],[25,20],[26,27],[27,28],[28,29],[29,30],[30,31],[31,26],[20,26],[21,27],[25,31]];

    const cubeVerts = [{x:-1,y:-1,z:-1},{x:1,y:-1,z:-1},{x:1,y:1,z:-1},{x:-1,y:1,z:-1},{x:-1,y:-1,z:1},{x:1,y:-1,z:1},{x:1,y:1,z:1},{x:-1,y:1,z:1}];
    const cubeEdges = [[0,1],[1,2],[2,3],[3,0],[4,5],[5,6],[6,7],[7,4],[0,4],[1,5],[2,6],[3,7]];

    const terminalVerts = [
        {x:-1.5, y:-1, z:0.4}, {x:1.5, y:-1, z:0.4}, {x:1.5, y:1, z:0.4}, {x:-1.5, y:1, z:0.4},
        {x:-1.5, y:-1, z:-0.2}, {x:1.5, y:-1, z:-0.2}, {x:1.5, y:1, z:-0.2}, {x:-1.5, y:1, z:-0.2},
        {x:-0.3, y:-1, z:0.1}, {x:0.3, y:-1, z:0.1}, {x:0.3, y:-1.5, z:0.1}, {x:-0.3, y:-1.5, z:0.1},
        {x:-0.8, y:-1.5, z:0.3}, {x:0.8, y:-1.5, z:0.3}, {x:0.8, y:-1.5, z:-0.3}, {x:-0.8, y:-1.5, z:-0.3},
        {x:-1.2, y:0.6, z:0.45}, {x:0.5, y:0.6, z:0.45}, {x:-1.2, y:0.2, z:0.45}, {x:1, y:0.2, z:0.45},
        {x:-1.2, y:-0.2, z:0.45}, {x:0, y:-0.2, z:0.45}, {x:-1.2, y:-0.6, z:0.45}, {x:0.8, y:-0.6, z:0.45}
    ];
    const terminalEdges = [[0,1],[1,2],[2,3],[3,0],[4,5],[5,6],[6,7],[7,4],[0,4],[1,5],[2,6],[3,7],[8,9],[9,10],[10,11],[11,8],[12,13],[13,14],[14,15],[15,12],[16,17],[18,19],[20,21],[22,23]];

    const potionVerts = [
        {x:-0.6, y:-1, z:0.3}, {x:0.6, y:-1, z:0.3}, {x:0.6, y:0.2, z:0.3}, {x:0.3, y:0.5, z:0.3},
        {x:0.3, y:0.8, z:0.3}, {x:-0.3, y:0.8, z:0.3}, {x:-0.3, y:0.5, z:0.3}, {x:-0.6, y:0.2, z:0.3},
        {x:-0.6, y:-1, z:-0.3}, {x:0.6, y:-1, z:-0.3}, {x:0.6, y:0.2, z:-0.3}, {x:0.3, y:0.5, z:-0.3},
        {x:0.3, y:0.8, z:-0.3}, {x:-0.3, y:0.8, z:-0.3}, {x:-0.3, y:0.5, z:-0.3}, {x:-0.6, y:0.2, z:-0.3},
        {x:-0.25, y:0.8, z:0.2}, {x:0.25, y:0.8, z:0.2}, {x:0.25, y:1.1, z:0.2}, {x:-0.25, y:1.1, z:0.2},
        {x:-0.25, y:0.8, z:-0.2}, {x:0.25, y:0.8, z:-0.2}, {x:0.25, y:1.1, z:-0.2}, {x:-0.25, y:1.1, z:-0.2}
    ];
    const potionEdges = [[0,1],[1,2],[2,3],[3,4],[4,5],[5,6],[6,7],[7,0],[8,9],[9,10],[10,11],[11,12],[12,13],[13,14],[14,15],[15,8],[0,8],[1,9],[2,10],[3,11],[4,12],[5,13],[6,14],[7,15],[16,17],[17,18],[18,19],[19,16],[20,21],[21,22],[22,23],[23,20],[16,20],[17,21],[18,22],[19,23]];

    function noise(x, y, t) {
        const n1 = Math.sin(x * 0.02 + t) * Math.cos(y * 0.02 - t * 0.5);
        const n2 = Math.sin(x * 0.015 - t * 0.7) * Math.sin(y * 0.025 + t * 0.3);
        const n3 = Math.cos(x * 0.01 + y * 0.01 + t * 0.2);
        return (n1 + n2 + n3) / 3;
    }

    function createGrid() {
        gridDots = [];
        const cols = Math.ceil(canvas.width / dotSpacing) + 1;
        const rows = Math.ceil(canvas.height / dotSpacing) + 1;
        for (let row = 0; row < rows; row++) {
            for (let col = 0; col < cols; col++) {
                gridDots.push({ baseX: col * dotSpacing, baseY: row * dotSpacing, x: col * dotSpacing, y: row * dotSpacing, col: col, row: row });
            }
        }
    }

    function resize() {
        canvas.width = window.innerWidth;
        canvas.height = window.innerHeight;
        createGrid();
    }

    window.addEventListener('resize', resize);
    resize();

    function rotatePoint(p, rx, ry) {
        let x = p.x * Math.cos(ry) - p.z * Math.sin(ry);
        let z = p.x * Math.sin(ry) + p.z * Math.cos(ry);
        let y = p.y * Math.cos(rx) - z * Math.sin(rx);
        z = p.y * Math.sin(rx) + z * Math.cos(rx);
        return { x: x, y: y, z: z };
    }

    function project(point, shape, w, h) {
        const fov = 400;
        const scale = fov / (fov + point.z + 300);
        return { x: (w * shape.x) + point.x * shape.scale * scale, y: (h * shape.y) + point.y * shape.scale * scale };
    }

    function drawGrid() {
        const centerX = canvas.width / 2;
        const centerY = canvas.height / 2;
        const maxDist = Math.sqrt(centerX * centerX + centerY * centerY);

        for (let i = 0; i < gridDots.length; i++) {
            const dot = gridDots[i];
            const distFromCenter = Math.sqrt(Math.pow(dot.baseX - centerX, 2) + Math.pow(dot.baseY - centerY, 2));
            const noiseVal = noise(dot.baseX, dot.baseY, time);
            const noiseVal2 = noise(dot.baseX * 1.5, dot.baseY * 1.5, time * 0.5);
            let baseSize = 1 + (noiseVal + 1) * 1.2;
            const radialFactor = 1 - (distFromCenter / maxDist);
            baseSize *= (0.5 + radialFactor * 0.8);
            let baseOpacity = 0.1 + (noiseVal2 + 1) * 0.15;
            baseOpacity *= (0.3 + radialFactor * 0.7);
            const wave1 = Math.sin(dot.baseX * 0.03 + time * 2) * 0.3;
            const wave2 = Math.cos(dot.baseY * 0.025 + time * 1.5) * 0.2;
            baseOpacity += (wave1 + wave2) * radialFactor * 0.15;
            let size = baseSize;
            let opacity = baseOpacity;
            let offsetX = 0;
            let offsetY = 0;

            if (mouse.x !== null && mouse.y !== null) {
            const dx = dot.baseX - mouse.x;
            const dy = dot.baseY - mouse.y;
            const distSq = dx * dx + dy * dy; // Faster multiplication
            const radSq = mouse.radius * mouse.radius;

            // Only calculate sqrt if the dot is actually close enough
            if (distSq < radSq) {
                const distance = Math.sqrt(distSq);
                const force = Math.pow((mouse.radius - distance) / mouse.radius, 1.5);
                    const angle = Math.atan2(dy, dx);
                    offsetX = Math.cos(angle) * force * 12;
                    offsetY = Math.sin(angle) * force * 12;
                    const ripple = Math.sin(distance * 0.05 - time * 5) * 0.5 + 0.5;
                    size = baseSize + force * 3 * ripple;
                    opacity = Math.min(1, opacity + force * 0.7);
                }
            }

            dot.x += ((dot.baseX + offsetX) - dot.x) * 0.2;
            dot.y += ((dot.baseY + offsetY) - dot.y) * 0.2;

            if (opacity > 0.02 && size > 0.3) {
                ctx.beginPath();
                ctx.arc(dot.x, dot.y, Math.max(0.5, size), 0, Math.PI * 2);
                ctx.fillStyle = themeColor;
                ctx.globalAlpha = Math.max(0, Math.min(1, opacity));
                ctx.fill();
            }
        }
        ctx.globalAlpha = 1;
    }

    function draw3DShapes() {
        ctx.strokeStyle = themeColor;
        ctx.lineWidth = 1.5;

        // COMMENT OUT or REMOVE these lines to disable shadows
        // ctx.shadowColor = themeColor;
        // ctx.shadowBlur = 10;
        
        ctx.globalAlpha = 0.55;

        shapes.forEach(function(shape) {
            shape.rx += shape.speedX;
            shape.ry += shape.speedY;
            shape.driftAngle += shape.driftSpeed;
            const driftX = Math.sin(shape.driftAngle) * shape.driftRadius;
            const driftY = Math.sin(shape.driftAngle * 0.7) * shape.driftRadius * 0.6;
            shape.baseX += shape.vx;
            shape.baseY += shape.vy;
            if (shape.baseX < 0.05 || shape.baseX > 0.95) { shape.vx *= -1; shape.baseX = Math.max(0.05, Math.min(0.95, shape.baseX)); }
            if (shape.baseY < 0.05 || shape.baseY > 0.95) { shape.vy *= -1; shape.baseY = Math.max(0.05, Math.min(0.95, shape.baseY)); }
            shape.x = shape.baseX + driftX;
            shape.y = shape.baseY + driftY;

            let verts, edges;
            if (shape.type === 'controller') { verts = controllerVerts; edges = controllerEdges; }
            else if (shape.type === 'codetag') { verts = codetagVerts; edges = codetagEdges; }
            else if (shape.type === 'cube') { verts = cubeVerts; edges = cubeEdges; }
            else if (shape.type === 'terminal') { verts = terminalVerts; edges = terminalEdges; }
            else if (shape.type === 'potion') { verts = potionVerts; edges = potionEdges; }
            if (!verts || !edges) return;

            ctx.beginPath();
            edges.forEach(function(edge) {
                const p1 = rotatePoint(verts[edge[0]], shape.rx, shape.ry);
                const p2 = rotatePoint(verts[edge[1]], shape.rx, shape.ry);
                const proj1 = project(p1, shape, canvas.width, canvas.height);
                const proj2 = project(p2, shape, canvas.width, canvas.height);
                ctx.moveTo(proj1.x, proj1.y);
                ctx.lineTo(proj2.x, proj2.y);
            });
            ctx.stroke();
        });
        ctx.shadowBlur = 0;
        ctx.globalAlpha = 1;
    }

    let lastFrameTime = 0;
    const fpsInterval = 1000 / 30; // Target 30 FPS

    function animate(currentTime) {
        // STOP if tab is not visible
        if (document.hidden) {
            requestAnimationFrame(animate); // Keep checking until visible again
            return; // Skip drawing and math
        }

        requestAnimationFrame(animate);

        const elapsed = currentTime - lastFrameTime;

        if (elapsed > fpsInterval) {
            lastFrameTime = currentTime - (elapsed % fpsInterval);
            
            // Your existing drawing code goes here
            time += 0.016; 
            ctx.clearRect(0, 0, canvas.width, canvas.height);
            drawGrid();
            draw3DShapes();
        }
    }
    requestAnimationFrame(animate);
}

// ==================== RENDER CONTENT SECTIONS ====================

function renderNavigation() {
    const navContainer = document.getElementById('main-nav');
    if (!navContainer || !navigation) return;

    let navHtml = '';
    navigation.forEach(item => {
        navHtml += `
        <a href="${item.link}" class="nav-item flex items-center gap-4 px-4 py-3 rounded-lg transition-all duration-300 group text-sm font-bold uppercase tracking-wider relative overflow-hidden text-gray-500 hover:text-white" data-section="${item.section}">
            <div class="nav-bg absolute inset-0 opacity-10 transition-transform duration-300 origin-left scale-x-0 group-hover:scale-x-100" style="background-color: var(--theme-color)"></div>
            <i data-lucide="${item.icon}" class="nav-icon w-[18px] h-[18px] relative z-10 transition-colors duration-300"></i>
            <span class="relative z-10">${item.label}</span>
            <i data-lucide="chevron-right" class="nav-chevron w-[14px] h-[14px] ml-auto relative z-10 animate-pulse hidden" style="color: var(--theme-color)"></i>
        </a>`;
    });

    navContainer.innerHTML = navHtml;
}

function renderSidebarSocials() {
    const container = document.getElementById('sidebar-socials');
    // Changed to use socialLinks directly to ensure we use iconType definitions
    const links = socialLinks; 
    
    if (!container || !links) return;

    let html = '';
    links.forEach(link => {
        if (link.visible === false) return;
        
        let iconElement = '';
        if (link.iconType === 'ionicon') {
            iconElement = `<ion-icon name="${link.icon}" class="w-5 h-5"></ion-icon>`;
        } else if (link.iconType === 'fontawesome') {
            // FIX: Use font-size and line-height for alignment instead of w/h
            iconElement = `<i class="fa-brands ${link.icon} text-xl leading-none"></i>`;
        } else {
            // Default fallback to Lucide
            iconElement = `<i data-lucide="${link.icon}" class="w-5 h-5"></i>`;
        }

        html += `
        <a href="${link.url}" target="_blank" class="flex items-center justify-center p-3 text-gray-500 hover:text-white hover:bg-white/10 rounded-full transition-all hover:scale-110" title="${link.platform}">
            ${iconElement}
        </a>`;
    });
    container.innerHTML = html;
    
    // Re-initialize Lucide icons in case any were added
    if(window.lucide) lucide.createIcons();
}

// Variable to store the timeout ID
let titleTypeTimeout = null;

function renderPersonal() {
    if (!siteConfig || !siteConfig.personal) return;

    const { personal } = siteConfig;
    
    // Update Sidebar Avatar & Name
    const avatarImg = document.getElementById('avatar-ring')?.querySelector('img');
    if (avatarImg) avatarImg.src = personal.avatar;
    
    const sidebarTitle = document.querySelector('#sidebar h1');
    if (sidebarTitle) sidebarTitle.textContent = personal.name;

    // --- Typewriter Effect ---
    const titleContainer = document.querySelector('#sidebar .flex.flex-wrap');
    
    if (titleContainer && personal.titles && personal.titles.length > 0) {
        titleContainer.innerHTML = '';
        const wrapper = document.createElement('div');
        wrapper.className = 'px-3 py-1.5 rounded bg-white/5 border border-white/10 inline-flex items-center justify-center min-w-[200px] min-h-[32px]';
        const typeText = document.createElement('span');
        typeText.className = 'text-[11px] font-mono text-gray-300 whitespace-nowrap';
        const cursor = document.createElement('span');
        cursor.textContent = '|';
        cursor.className = 'ml-1 animate-pulse text-[11px]';
        cursor.style.color = 'var(--theme-color)';
        wrapper.appendChild(typeText);
        wrapper.appendChild(cursor);
        titleContainer.appendChild(wrapper);

        if (titleTypeTimeout) clearTimeout(titleTypeTimeout);

        let loopNum = 0;
        let isDeleting = false;
        let charIndex = 0;
        let typeSpeed = 100;

        function type() {
            const i = loopNum % personal.titles.length;
            const fullTxtArray = Array.from(personal.titles[i]);

            if (isDeleting) {
                charIndex--;
                typeSpeed = 50; 
            } else {
                charIndex++;
                typeSpeed = 100; 
            }

            typeText.textContent = fullTxtArray.slice(0, charIndex).join('');

            if (!isDeleting && charIndex === fullTxtArray.length) {
                typeSpeed = 2000; 
                isDeleting = true;
            } else if (isDeleting && charIndex === 0) {
                isDeleting = false;
                loopNum++;
                typeSpeed = 500; 
            }

            titleTypeTimeout = setTimeout(type, typeSpeed);
        }
        type();
    }
    
    // Update Hero Section
    if (personal.about) {
        const heroTitle = document.getElementById('hero-title');
        const heroDesc = document.getElementById('hero-desc');
        if (heroTitle) heroTitle.innerHTML = personal.about.greeting;
        if (heroDesc) heroDesc.innerHTML = personal.about.description;
    }

    // Update Hero Buttons - REDUCED GLOW HERE
    if (siteConfig.overview_buttons) {
        const btnContainer = document.getElementById('hero-buttons');
        if (btnContainer) {
            let btnsHtml = '';
            siteConfig.overview_buttons.forEach(btn => {
                if (btn.visible === false) return;
                
                if (btn.style === 'primary') {
                    // Reduced shadow from 30px to 10px
                    btnsHtml += `<a href="${btn.href}" id="cta-primary" class="px-8 py-4 text-black font-black uppercase tracking-widest text-sm rounded transition-all duration-300 hover:scale-105 hover:-translate-y-1" style="background-color: var(--theme-color); box-shadow: 0 0 10px var(--theme-color);">${btn.label}</a>`;
                } else {
                    btnsHtml += `<a href="${btn.href}" class="px-8 py-4 border-2 border-white/20 text-white font-bold uppercase tracking-widest text-sm rounded hover:bg-white/10 hover:border-white transition-colors backdrop-blur-sm">${btn.label}</a>`;
                }
            });
            btnContainer.innerHTML = btnsHtml;
        }
    }
}

function renderFooter() {
    const footerEl = document.getElementById('main-footer');
    if (!footerEl || !footerConfig || !footerConfig.enabled) {
        if(footerEl) footerEl.style.display = 'none';
        return;
    }

    let html = '';

    

    if (footerConfig.text) {
        html += `<p>${footerConfig.text}</p>`;
    }
    if (footerConfig.copyright && footerConfig.copyright.enabled) {
        html += `<p class="mt-2">${footerConfig.copyright.text}</p>`;
    }

    // --- UPDATED: Visit Counter Logic ---
    if (footerConfig.visitCounter) {
        // Changed page_id to 'kenanege.com' to track your custom domain
        html += `<br> <div class="mb-6 flex justify-center">
            <img src="https://visitor-badge.laobi.icu/badge?page_id=kenanege.com&left_color=black&right_color=bc13fe&left_text=Total Visit:" alt="Total Visit:" style="border-radius: 4px; box-shadow: 0 0 10px rgba(188, 19, 254, 0.3);">
        </div>`;
    }
    // ------------------------------------
    
    footerEl.innerHTML = html;
}

// ==================== HELPER FUNCTIONS ====================
function capitalizeCategory(category) {
    return category.split(' ').map(word => word.charAt(0).toUpperCase() + word.slice(1)).join(' ');
}

function getHoverIcon(iconType) {
    const iconMap = {
        'unity': 'gamepad-2', 'unreal': 'gamepad-2', 'web': 'globe', 'django': 'server',
        'python': 'terminal', 'cpp': 'code', 'csharp': 'code', 'android': 'smartphone',
        'design': 'palette', 'music': 'music', 'default': 'eye'
    };
    return iconMap[iconType] || iconMap['default'];
}

function renderGames() {
    const grid = document.getElementById('games-grid');
    if (!grid) return;

    const highlightedGames = gamesShowcase.highlighted ? gamesShowcase.highlighted.filter(g => g.visible !== false) : [];

    if (highlightedGames.length === 0) return;

    let html = '';
    // Reduced text shadow here
    highlightedGames.forEach((game, idx) => {
        // CHANGED: Use special classes only on desktop (md:), default to aspect-video for mobile
        const sizeClass = idx === 0 ? 'aspect-video md:col-span-2 md:aspect-[21/9]' : 'aspect-video';
        
        html += `<a href="${game.link}" target="_blank" class="game-card group relative ${sizeClass} bg-black/80 border border-white/10 rounded-xl overflow-hidden transition-all duration-300 hover:shadow-xl">
            <div class="game-border absolute inset-0 border-2 border-transparent transition-colors duration-300 rounded-xl z-20 pointer-events-none"></div>
            <div class="game-bg absolute inset-0 bg-cover bg-center opacity-60 transition-all duration-700" style="background-image: url(${game.image})"></div>
            <div class="absolute inset-0 bg-gradient-to-t from-black via-transparent to-transparent opacity-90"></div>
            <div class="absolute bottom-0 left-0 p-6 z-10">
                <h3 class="text-2xl md:text-4xl font-black uppercase italic text-white mb-2 drop-shadow-md">${game.title}</h3>
                <div class="flex items-center gap-2 font-mono text-sm uppercase tracking-widest transition-colors duration-300" style="color: var(--theme-color); text-shadow: 0 0 5px var(--theme-color)">
                    <i data-lucide="gamepad-2" class="w-4 h-4"></i> Click to Play
                </div>
            </div>
        </a>`;
    });
    grid.innerHTML = html;
}

function renderPortfolio(filter) {
    filter = filter || 'all';
    let items = portfolioItems;

    if (filter === 'game') items = items.filter(i => i.category.includes('game'));
    else if (filter === 'web') items = items.filter(i => i.category.includes('web'));
    else if (filter === 'android') items = items.filter(i => i.category.includes('mobile'));
    else if (filter === 'windows') items = items.filter(i => i.category.includes('windows'));
    else if (filter === 'other') items = items.filter(i => i.category.includes('other'));
    else if (filter === 'highlighted') items = items.filter(i => i.highlighted === true);
    else if (filter === 'awards') items = items.filter(i => i.award);

    const grid = document.getElementById('portfolio-grid');
    if (!grid) return;

    let html = '';
    items.forEach(item => {
        const iconName = getHoverIcon(item.iconType);
        const targetAttr = (item.opennewtab !== false) ? 'target="_blank" rel="noopener noreferrer"' : '';

        let tagsHtml = '';
        if (item.tags && item.tags.length > 0) {
            tagsHtml = item.tags.slice(0, 3).map((tag, index) => {
                let styleAttr = '';
                let tagClass = "bg-black/90 text-white border-white/10"; // Darker tag bg
                
                if (index === 2) {
                    tagClass = "text-black font-bold border-transparent";
                    // Reduced shadow from 10px to 5px
                    styleAttr = `style="background-color: var(--theme-color); box-shadow: 0 0 5px var(--theme-color);"`;
                }

                return `<span class="px-2 py-1 text-[8px] uppercase font-bold border shadow-[0_0_5px_black] backdrop-blur-sm ${tagClass}" ${styleAttr}>${tag}</span>`;
            }).join('');
        }

        // Updated Card Background to be darker: bg-[#050505]/90
        html += `<a href="${item.link}" ${targetAttr} class="project-card group relative block bg-[#050505]/90 backdrop-blur-md border border-white/10 rounded-lg h-full flex flex-col overflow-hidden transition-all duration-300">
            <div class="card-border absolute inset-0 border border-transparent transition-colors duration-300 pointer-events-none rounded-lg z-20"></div>
            <div class="aspect-video w-full bg-[#050505] relative overflow-hidden border-b border-white/5">
                <img src="${item.image}" loading="lazy" alt="${item.title}" class="card-image w-full h-full object-cover transition-transform duration-700 opacity-80" onerror="this.style.display='none'; this.nextElementSibling.style.display='flex'">
                <div class="w-full h-full items-center justify-center bg-[#0a0a0a] relative hidden" style="color: var(--theme-color)">
                    <i data-lucide="gamepad-2" class="w-10 h-10 opacity-50 relative z-10 animate-pulse"></i>
                </div>
                <div class="absolute top-2 right-2 flex flex-col items-end gap-1">${tagsHtml}</div>
            </div>
            <div class="p-5 flex-grow flex flex-col relative z-10">
                <h4 class="card-title text-lg font-bold text-white mb-2 leading-tight transition-colors">${item.title}</h4>
                <p class="text-gray-400 text-xs mb-4 flex-grow font-mono leading-relaxed group-hover:text-gray-300">${item.description}${item.award ? '<br><span class="text-yellow-400 font-semibold"> ' + item.award + '</span>' : ''}</p>
                <div class="flex items-center justify-between mt-auto pt-4 border-t border-white/5 font-bold text-[10px] uppercase tracking-wider transition-colors duration-300" style="color: var(--theme-color)">
                    <span class="flex items-center gap-2">
                        <i data-lucide="${iconName}" class="w-3 h-3"></i>
                        ${capitalizeCategory(item.category)}
                    </span>
                    <i data-lucide="external-link" class="w-3 h-3 group-hover:translate-x-1 transition-transform"></i>
                </div>
            </div>
        </a>`;
    });
    grid.innerHTML = html;
    
    if(window.lucide) lucide.createIcons();
}

function renderExperience() {
    const list = document.getElementById('experience-list');
    if (!list || experience.length === 0) return;

    // Clear previous classes to prevent conflicts
    list.className = '';

    // --- CARD GENERATOR HELPER ---
    // Creates a flexible height card. 
    // We removed 'h-full' so it shrinks to fit content exactly.
    const createCard = (job) => {
        let positionsHtml = '';
        job.positions.forEach((pos, index) => {
            let responsibilitiesHtml = '';
            if (pos.responsibilities && pos.responsibilities.length > 0) {
                responsibilitiesHtml = `<ul class="mt-3 space-y-2">`;
                pos.responsibilities.forEach(r => {
                    responsibilitiesHtml += `<li class="text-sm text-gray-300 leading-relaxed flex items-start gap-2"><span class="text-[10px] mt-1.5" style="color: var(--theme-color)">▸</span><span>${r}</span></li>`;
                });
                responsibilitiesHtml += `</ul>`;
            }
            
            // Separator between multiple positions
            const separator = index > 0 ? '<div class="h-px bg-white/10 my-4"></div>' : '';

            positionsHtml += `${separator}
            <div class="relative">
                <div class="flex flex-col sm:flex-row sm:justify-between sm:items-baseline mb-2">
                    <h4 class="text-lg font-bold text-white transition-colors group-hover:text-[var(--theme-color)]">${pos.title}</h4>
                    <span class="text-xs font-mono text-gray-400 bg-white/5 px-2 py-1 rounded border border-white/5 whitespace-nowrap mt-1 sm:mt-0">${pos.startDate} — ${pos.endDate}</span>
                </div>
                ${responsibilitiesHtml}
            </div>`;
        });

        // Company Logo Logic
        const logoHtml = (job.logo && job.url)
            ? `<a href="${job.url}" target="_blank" class="shrink-0 p-2 bg-white/5 rounded-xl border border-white/5 hover:border-[var(--theme-color)] transition-colors"><img src="${job.logo}" loading="lazy" alt="${job.company} logo" class="w-10 h-10 object-contain"></a>`
            : '';

        // Note: Removed 'h-full' class to ensure height fits content
        return `<div class="group bg-black/70 backdrop-blur-md p-6 rounded-2xl border border-white/10 hover:border-[var(--theme-color)] transition-all duration-300 hover:-translate-y-1 hover:shadow-lg flex flex-col w-full">
            <div class="flex items-center gap-4 mb-6 border-b border-white/10 pb-4">
                ${logoHtml}
                <div>
                    <h3 class="text-2xl font-bold text-white leading-none mb-1">${job.company}</h3>
                    ${job.url ? `<a href="${job.url}" target="_blank" class="text-xs font-mono opacity-60 hover:opacity-100 hover:text-[var(--theme-color)] transition-all flex items-center gap-1">Visit Website <i data-lucide="external-link" class="w-3 h-3"></i></a>` : ''}
                </div>
            </div>
            <div class="flex-grow">
                ${positionsHtml}
            </div>
        </div>`;
    };

    // --- MOBILE LAYOUT (Linear) ---
    // Simple stack for mobile so chronological order is strictly preserved (1, 2, 3...)
    const mobileHtml = `
        <div class="flex flex-col gap-6 md:hidden">
            ${experience.map(job => createCard(job)).join('')}
        </div>
    `;

    // --- DESKTOP LAYOUT (Smart Masonry) ---
    // Calculates approximate height of each card (based on bullet points)
    // and distributes them to the shorter column to keep layout balanced.
    const leftItems = [];
    const rightItems = [];
    let leftHeight = 0;
    let rightHeight = 0;

    experience.forEach(job => {
        // Estimate height: base padding (100) + title (50) + bullets (50 each)
        // This is a rough heuristic to balance the columns visually
        let jobHeight = 100; 
        job.positions.forEach(pos => {
            jobHeight += 50; // Title row
            if (pos.responsibilities) {
                jobHeight += pos.responsibilities.length * 50; // ~50px per bullet point including margin
            }
        });

        // Add to the shorter column to balance height
        // NOTE: We favor left column slightly if equal (<=)
        if (leftHeight <= rightHeight) {
            leftItems.push(job);
            leftHeight += jobHeight;
        } else {
            rightItems.push(job);
            rightHeight += jobHeight;
        }
    });

    const leftColHtml = leftItems.map(job => createCard(job)).join('<div class="h-6"></div>'); // Manual gap or flex gap
    const rightColHtml = rightItems.map(job => createCard(job)).join('<div class="h-6"></div>');

    // Using 'items-start' ensures columns don't stretch to match each other's height
    const desktopHtml = `
        <div class="hidden md:grid grid-cols-2 gap-6 items-start">
            <div class="flex flex-col gap-6">
                ${leftItems.map(job => createCard(job)).join('')}
            </div>
            <div class="flex flex-col gap-6">
                ${rightItems.map(job => createCard(job)).join('')}
            </div>
        </div>
    `;

    list.innerHTML = mobileHtml + desktopHtml;
    
    // Re-initialize Lucide icons for the newly added elements
    if(window.lucide) lucide.createIcons();
}

function renderEducation() {
    const grid = document.getElementById('education-grid');
    if (!grid || education.length === 0) return;

    let html = '';
    education.forEach(edu => {
        let degreesHtml = '';
        if (edu.degrees && edu.degrees.length > 0) {
            edu.degrees.forEach(degree => {
                degreesHtml += `<div class="mb-3 last:mb-0">
                    <h5 class="text-sm font-bold text-white">${degree.title}</h5>
                    <span class="text-xs text-gray-500 font-mono">${degree.startDate} — ${degree.endDate}</span>
                    ${degree.description ? `<p class="text-xs text-gray-400 mt-1">${degree.description}</p>` : ''}
                </div>`;
            });
        }

        const logoHtml = edu.logo
            ? `<img src="${edu.logo}" loading="lazy" alt="${edu.institution} logo" class="w-10 h-10 object-contain">`
            : `<i data-lucide="graduation-cap" class="w-8 h-8"></i>`;

        const clickableClass = edu.url ? 'cursor-pointer hover:scale-[1.02]' : '';
        const onClickAttr = edu.url ? `onclick="window.open('${edu.url}', '_blank')"` : '';

        // Darker BG: bg-black/70
        html += `<div class="edu-card group p-6 bg-black/70 backdrop-blur-md border border-white/10 rounded-2xl transition-all duration-300 ${clickableClass}" ${onClickAttr}>
            <div class="flex items-start gap-4 mb-4">
                <div class="edu-icon transition-colors duration-300 p-2 bg-white/5 rounded-lg" style="color: var(--theme-color)">
                    ${logoHtml}
                </div>
                <div class="flex-grow">
                    <h4 class="edu-title text-lg font-bold text-white leading-tight mb-1 transition-colors">${edu.institution}</h4>
                    ${edu.url ? `<span class="text-xs font-mono opacity-50 group-hover:opacity-100 transition-opacity" style="color: var(--theme-color)">Click to visit →</span>` : ''}
                </div>
            </div>
            <div class="border-t border-white/5 pt-4">${degreesHtml}</div>
        </div>`;
    });
    grid.innerHTML = html;
}

/**
 * Robustly fetches SVG widgets with fallback, timeout, and content validation.
 */
async function loadSvgWidget(url, imgId, fallbackUrl) {
    const imgElement = document.getElementById(imgId);
    if (!imgElement) return;

    // 1. Show Fallback IMMEDIATELY
    // This ensures the user sees the static image while the dynamic one loads.
    imgElement.src = fallbackUrl;
    imgElement.style.opacity = "0.8"; // Indicate loading/fallback state
    imgElement.style.borderRadius = "10px";

    try {
        // 2. Setup a timeout (e.g., 5 seconds)
        // If the external API is slow or hanging, we abort to save resources.
        const controller = new AbortController();
        const timeoutId = setTimeout(() => controller.abort(), 5000);

        // 3. Fetch with specific headers
        const response = await fetch(url, { 
            signal: controller.signal,
            cache: 'no-cache',      // Ensure we don't serve cached error states
            headers: { 'Accept': 'image/svg+xml' } 
        });
        
        clearTimeout(timeoutId); // Clear timeout if fetch completes

        // 4. Check HTTP Status
        if (!response.ok) throw new Error(`HTTP Error ${response.status}`);
        
        // 5. Get text content to validate
        const svgText = await response.text();
        
        // 6. VALIDATION: Check content integrity
        // Must start with SVG tag (ignoring whitespace)
        if (!svgText.trim().startsWith('<svg')) {
            throw new Error("Response is not a valid SVG");
        }

        // Check for specific error phrases often returned by stats widgets
        // These widgets often return a 200 OK but with an SVG containing text like "Something went wrong"
        const errorKeywords = [
            "Something went wrong",
            "could not resolve",
            "Failed to fetch",
            "banned",
            "rate limit",
            "bad request",
            "internal server error"
        ];
        
        // Case-insensitive check
        const lowerSvg = svgText.toLowerCase();
        if (errorKeywords.some(key => lowerSvg.includes(key.toLowerCase()))) {
            throw new Error("SVG contains visible error message");
        }

        // 7. Success: Render the dynamic SVG
        const blob = new Blob([svgText], {type: 'image/svg+xml'});
        const blobUrl = URL.createObjectURL(blob);
        
        // Clean up previous blob if exists (optional but good for memory)
        if (imgElement.dataset.blobUrl) {
            URL.revokeObjectURL(imgElement.dataset.blobUrl);
        }
        
        imgElement.onload = () => {
            imgElement.style.opacity = "1"; // Fade in full opacity
        };
        
        imgElement.src = blobUrl;
        imgElement.dataset.blobUrl = blobUrl; // Store reference for cleanup

    } catch (error) {
        console.warn(`GitHub Widget (${imgId}) kept fallback. Reason:`, error.message);
        // Fallback is already displayed from step 1, so no action needed.
    }
}

function renderMore() {
    if (!siteConfig) return;

    // Spotify
    const spotifyEmbed = document.getElementById('spotify-embed');
    if (spotifyEmbed && siteConfig.external && siteConfig.external.spotifyArtistId) {
        spotifyEmbed.src = 'https://open.spotify.com/embed/artist/' + siteConfig.external.spotifyArtistId + '?utm_source=generator&theme=0';
    }

    // Steam
    const steamStats = document.getElementById('steam-stats');
    const steamProfileLink = document.getElementById('steam-profile-link');
    if (siteConfig.external && siteConfig.external.steamUsername) {
        if (steamProfileLink) steamProfileLink.href = 'https://steamcommunity.com/id/' + siteConfig.external.steamUsername;
        if (steamStats) steamStats.src = 'https://steam-stat.vercel.app/api?profileName=' + siteConfig.external.steamUsername;
    }

    // GitHub Section
    if (siteConfig.external && siteConfig.external.githubUsername) {
        const ghUser = siteConfig.external.githubUsername;
        
        // Path to your snake animation fallback
        const fallbackImageStats = './assets/images/fallback/streak.svg';
        const fallbackImageStreak = './assets/images/fallback/stats.svg';  

        // URLs for the widgets
        const statsUrl = `https://github-readme-stats.vercel.app/api?username=${ghUser}&show_icons=true&theme=github_dark&hide_border=true&bg_color=0d1117`;
        const streakUrl = `https://github-readme-streak-stats.herokuapp.com/?user=${ghUser}&theme=github-dark-blue&hide_border=true&background=0d1117`;

        // Execute the safe loader
        loadSvgWidget(statsUrl, 'github-stats', fallbackImageStats);
        loadSvgWidget(streakUrl, 'github-streak', fallbackImageStreak);
    }

    // Instagram
    if (instagram) {
        const instaAvatar = document.getElementById('insta-avatar');
        const instaUsername = document.getElementById('insta-username');
        const instaGrid = document.getElementById('insta-grid');

        if (instaAvatar && instagram.profileImage) instaAvatar.src = instagram.profileImage;
        if (instaUsername && instagram.username) {
            instaUsername.href = instagram.profileUrl;
            instaUsername.textContent = '@' + instagram.username;
        }

        if (instaGrid && instagram.posts && instagram.posts.length > 0) {
            let instaHtml = '';
            instagram.posts.forEach(post => {
                instaHtml += `<a href="${post.link}" target="_blank" class="block aspect-square overflow-hidden rounded-lg group">
                    <img src="${post.image}" loading="lazy" alt="${post.alt}" class="w-full h-full object-cover transition-transform duration-300 group-hover:scale-110">
                </a>`;
            });
            instaGrid.innerHTML = instaHtml;
        }
    }
}

// ==================== THEME & UI LOGIC ====================

function updateTheme(section) {
    currentSection = section;
    const targetHex = (themes[section] && themes[section].hex) ? themes[section].hex : '#bc13fe';
    
    // Parse new target and start transition
    const target = hexToRgb(targetHex);
    if(target) {
        targetThemeRgb = target;
        if(!isTransitioningColor) {
            isTransitioningColor = true;
            processColorTransition();
        }
    }

    // Update Nav Active States
    document.querySelectorAll('.nav-item').forEach(item => {
        const isActive = item.dataset.section === section;
        const chevron = item.querySelector('.nav-chevron');
        const icon = item.querySelector('.nav-icon');

        if (isActive) {
            item.classList.add('active', 'text-white');
            item.classList.remove('text-gray-500');
            if (chevron) chevron.classList.remove('hidden');
            // Allow CSS var to handle color transition naturally via stylesheet
            if (icon) icon.style.color = ''; 
        } else {
            item.classList.remove('active', 'text-white');
            item.classList.add('text-gray-500');
            if (chevron) chevron.classList.add('hidden');
            if (icon) icon.style.color = ''; 
        }
    });

    updateFilterStyles();
}

function updateDynamicColors() {
    const elementsToColor = [
        { selector: '#status-badge', property: 'borderColor', value: themeColor + '60' },
        { selector: '#hero-accent', property: 'color', value: themeColor },
        { selector: '#avatar-ring', property: 'borderColor', value: themeColor },
        { selector: '#cta-primary', property: 'backgroundColor', value: themeColor },
        { selector: '[data-lucide="map-pin"]', property: 'color', value: themeColor },
        { selector: '#sidebar-glow', property: 'background', value: `linear-gradient(to right, transparent, ${themeColor}, transparent)` }
    ];

    elementsToColor.forEach(el => {
        const domEl = document.querySelector(el.selector);
        if (domEl) domEl.style[el.property] = el.value;
    });

    // REDUCED GLOW LOGIC
    const heroAccent = document.getElementById('hero-accent');
    if (heroAccent) {
        // Significantly reduced shadow for cleaner look, especially on mobile
        heroAccent.style.textShadow = `0 0 8px ${themeColor}`;
    }
    
    // Reduce avatar glow
    const avatarRing = document.getElementById('avatar-ring');
    if (avatarRing) {
        avatarRing.style.boxShadow = `0 0 10px ${themeColor}`; // Reduced from 20px
    }

    document.querySelectorAll('.nav-item.active .nav-icon').forEach(icon => {
        icon.style.color = themeColor;
    });
}

function updateFilterStyles() {
    document.querySelectorAll('.filter-btn').forEach(btn => {
        const isActive = btn.dataset.filter === currentFilter;
        if (isActive) {
            btn.style.backgroundColor = themeColor;
            btn.style.borderColor = themeColor;
            btn.style.color = 'black';
        } else {
            btn.style.backgroundColor = 'transparent';
            btn.style.borderColor = 'rgba(255,255,255,0.2)';
            btn.style.color = '#9ca3af';
        }
    });
}

function setupCVLocalization() {
    const cvLinks = [document.getElementById('cv-download-link'), document.getElementById('cv-download-link-exp')];
    if (!siteConfig || !siteConfig.cv) return;

    const userInTurkey = () => {
        const lang = (navigator.language || '').toLowerCase();
        const tz = (Intl.DateTimeFormat().resolvedOptions().timeZone || '').toLowerCase();
        return lang.startsWith('tr') || tz.includes('istanbul');
    };

    const cvUrl = userInTurkey() ? siteConfig.cv.turkey : siteConfig.cv.default;

    cvLinks.forEach(link => {
        if (link) link.href = cvUrl;
    });
}

function initEventListeners() {
    let scrollThrottle;
    window.addEventListener('scroll', function() {
        if (!scrollThrottle) {
            scrollThrottle = setTimeout(function() {
                scrollThrottle = null;
                const sections = navigation.map(n => n.section);
                const scrollPos = window.scrollY + window.innerHeight * 0.3;
                
                for (const section of sections) {
                    const el = document.getElementById(section);
                    if (el && scrollPos >= el.offsetTop && scrollPos < el.offsetTop + el.offsetHeight) {
                        if (currentSection !== section) updateTheme(section);
                        break;
                    }
                }
            }, 50);
        }
    });

    // Find this block inside the initEventListeners() function
    document.getElementById('main-nav').addEventListener('click', function(e) {
        const link = e.target.closest('.nav-item');
        if (!link) return;
        e.preventDefault();
        const section = link.dataset.section;
        const el = document.getElementById(section);
        
        // --- MODIFIED SECTION START ---
        if (el) {
            const offset = 40; // Adjust this value (in pixels) to change the gap size
            const elementPosition = el.getBoundingClientRect().top + window.scrollY;
            const offsetPosition = elementPosition - offset;

            window.scrollTo({
                top: offsetPosition,
                behavior: 'smooth'
            });
        }
        // --- MODIFIED SECTION END ---
        
        document.getElementById('sidebar').classList.add('-translate-x-full');
        document.getElementById('menu-icon').classList.remove('hidden');
        document.getElementById('close-icon').classList.add('hidden');
    });
    
    document.getElementById('mobile-toggle').addEventListener('click', function() {
        const sidebar = document.getElementById('sidebar');
        const isOpen = !sidebar.classList.contains('-translate-x-full');
        if (isOpen) {
            sidebar.classList.add('-translate-x-full');
            document.getElementById('menu-icon').classList.remove('hidden');
            document.getElementById('close-icon').classList.add('hidden');
        } else {
            sidebar.classList.remove('-translate-x-full');
            document.getElementById('menu-icon').classList.add('hidden');
            document.getElementById('close-icon').classList.remove('hidden');
        }
    });

    document.querySelectorAll('.filter-btn').forEach(function(btn) {
        btn.addEventListener('click', function() {
            currentFilter = btn.dataset.filter;
            updateFilterStyles();
            renderPortfolio(currentFilter);
        });
    });
}

// ==================== INIT ====================
async function init() {
    await loadData();
    
    renderNavigation();
    renderPersonal();
    renderSidebarSocials();
    renderGames();
    renderPortfolio('highlighted');
    renderExperience();
    renderEducation();
    renderMore();
    renderFooter();
    
    setupCVLocalization();
    initEventListeners();
    
    lucide.createIcons();
    initBackground();
    
    const hash = window.location.hash.substring(1); // Remove '#'
    const defaultNav = navigation.find(n => n.default);
    let startSection = (hash && document.getElementById(hash)) ? hash : (defaultNav ? defaultNav.section : 'portfolio');
    
    updateTheme(startSection);
    updateFilterStyles();

    setTimeout(() => {
        const el = document.getElementById(startSection);
        if(el) {
            const offset = 40; // Ensure this matches the gap size you set in the click listener
            const elementPosition = el.getBoundingClientRect().top + window.scrollY;
            const offsetPosition = elementPosition - offset;

            window.scrollTo({
                top: offsetPosition,
                behavior: 'smooth' 
            });
        }
    }, 100);
}

function styleNominationTags() {
    document.querySelectorAll('#portfolio-grid span').forEach(el => {
        if (el.textContent.trim().startsWith('🏆')) {
            el.classList.add('nominee-tag');
        }
    });
}
setTimeout(styleNominationTags, 500);
document.addEventListener('click', (e) => {
    if (e.target.classList.contains('filter-btn')) {
        setTimeout(styleNominationTags, 150);
    }
});



document.addEventListener('DOMContentLoaded', init);

