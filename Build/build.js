const fs = require('fs');
const path = require('path');
const ejs = require('ejs');

// ── Paths ───────────────────────────────────────────────────────────────────
const ROOT = path.resolve(__dirname, '..');
const DIST = path.join(__dirname, 'dist');
const TEMPLATES = path.join(__dirname, 'templates');
const DATA_DIR = path.join(ROOT, 'data');
const STATIC_SRC = path.join(ROOT, 'web', 'static');
const CONTENT_DIR = path.join(DATA_DIR, 'content');

// ── Helpers ─────────────────────────────────────────────────────────────────
function ensureDir(dir) {
    fs.mkdirSync(dir, { recursive: true });
}

function copyDirSync(src, dest) {
    ensureDir(dest);
    for (const entry of fs.readdirSync(src, { withFileTypes: true })) {
        const srcPath = path.join(src, entry.name);
        const destPath = path.join(dest, entry.name);
        if (entry.isDirectory()) {
            copyDirSync(srcPath, destPath);
        } else {
            fs.copyFileSync(srcPath, destPath);
        }
    }
}

function rootPrefix(depth) {
    // For web server deployment, use empty string (absolute paths like /static/...)
    // For local file:// testing, uncomment the relative version below:
    // if (depth === 0) return '.';
    // return Array(depth).fill('..').join('/');
    return '';
}

function renderPage(templateFile, data) {
    const templatePath = path.join(TEMPLATES, 'pages', templateFile);
    const pageContent = ejs.render(fs.readFileSync(templatePath, 'utf8'), data, {
        filename: templatePath
    });

    const layoutPath = path.join(TEMPLATES, 'layouts', 'base.ejs');
    return ejs.render(fs.readFileSync(layoutPath, 'utf8'), {
        ...data,
        content: pageContent
    }, {
        filename: layoutPath
    });
}

function writePage(outputPath, html) {
    ensureDir(path.dirname(outputPath));
    fs.writeFileSync(outputPath, html, 'utf8');
    console.log('  -> ' + path.relative(DIST, outputPath));
}

// ── Load Data ───────────────────────────────────────────────────────────────
console.log('Loading data...');
const projectsData = JSON.parse(fs.readFileSync(path.join(DATA_DIR, 'projects.json'), 'utf8'));
const filterTags = JSON.parse(fs.readFileSync(path.join(DATA_DIR, 'filter-tags.json'), 'utf8'));

// Deduplicate projects (same logic as Go AllProjects)
const seen = new Set();
const allProjects = [];
for (const section of projectsData.sections) {
    for (const p of section.projects) {
        if (!seen.has(p.slug)) {
            seen.add(p.slug);
            allProjects.push(p);
        }
    }
}

// Group by year, descending
const yearMap = {};
for (const p of allProjects) {
    if (!yearMap[p.year]) yearMap[p.year] = [];
    yearMap[p.year].push(p);
}
const yearGroups = Object.keys(yearMap)
    .map(Number)
    .sort((a, b) => b - a)
    .map(year => ({ year, projects: yearMap[year] }));

console.log(`Found ${allProjects.length} unique projects in ${yearGroups.length} year groups`);

// ── Clean dist ──────────────────────────────────────────────────────────────
if (fs.existsSync(DIST)) {
    fs.rmSync(DIST, { recursive: true });
}
ensureDir(DIST);

// ── Build Pages ─────────────────────────────────────────────────────────────
console.log('\nBuilding pages...');

// Home page (index.html) — depth 0
writePage(
    path.join(DIST, 'index.html'),
    renderPage('home.ejs', {
        title: 'Portfolio',
        root: rootPrefix(0),
        filterTags,
        yearGroups
    })
);

// Resume page — depth 1 (resume/index.html)
writePage(
    path.join(DIST, 'resume', 'index.html'),
    renderPage('resume.ejs', {
        title: 'Resume',
        root: rootPrefix(1),
        headExtra: '<link rel="stylesheet" href="' + rootPrefix(1) + '/static/css/resume.css">'
    })
);

// Contact page — depth 1 (contact/index.html)
writePage(
    path.join(DIST, 'contact', 'index.html'),
    renderPage('contact.ejs', {
        title: 'Contact',
        root: rootPrefix(1)
    })
);

// Project detail pages (skip external-URL projects) — depth 2 (portfolio/slug/index.html)
for (const project of allProjects) {
    if (project.externalUrl) {
        console.log(`  -- skipping ${project.slug} (external URL)`);
        continue;
    }

    // Load custom content if available
    let content = null;
    const contentPath = path.join(CONTENT_DIR, project.slug + '.html');
    if (fs.existsSync(contentPath)) {
        content = fs.readFileSync(contentPath, 'utf8');
    }

    writePage(
        path.join(DIST, 'portfolio', project.slug, 'index.html'),
        renderPage('project.ejs', {
            title: project.title,
            root: rootPrefix(2),
            project,
            content
        })
    );
}

// ── Copy Static Assets ──────────────────────────────────────────────────────
console.log('\nCopying static assets...');
copyDirSync(STATIC_SRC, path.join(DIST, 'static'));

// Copy favicon to root
const faviconSrc = path.join(STATIC_SRC, 'favicon.svg');
if (fs.existsSync(faviconSrc)) {
    fs.copyFileSync(faviconSrc, path.join(DIST, 'favicon.svg'));
    console.log('  -> favicon.svg');
}

console.log('\nBuild complete! Output in Build/dist/');
