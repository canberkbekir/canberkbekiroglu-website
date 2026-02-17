// === Utility Functions ===
const clamp = (a, min = 0, max = 1) => Math.min(max, Math.max(min, a));
const invlerp = (x, y, a) => (a - x) / (y - x);
function lerp(start, end, amt) {
    return (1 - amt) * start + amt * end;
}

// === Parallax Effects ===
function updateParallaxLayerSize() {
    var elements = document.getElementsByClassName("parallax-layer");
    for (let i = 0; i < elements.length; i++) {
        var element = elements[i];
        var containerRect = element.parentElement.parentElement.getBoundingClientRect();
        element.style.width = containerRect.width + "px";
        element.style.height = containerRect.height + "px";
    }
}

function updateRelativeElements() {
    var elements = document.getElementsByClassName("parallax-focalanchortop");
    for (let i = 0; i < elements.length; i++) {
        var element = elements[i];
        var clientRect = element.getBoundingClientRect();
        var time = clamp(invlerp(clientRect.height, -clientRect.height, clientRect.top), 0, 1);
        element.style.setProperty("--animation-time", time.toFixed(5));
        element.style.setProperty("--parallax-offset", lerp(-clientRect.height, clientRect.height, time).toFixed(1));
    }
}

// === Navigation & Sidebar TOC Highlighting ===
function navHighlighter() {
    let sections = document.querySelectorAll("section[id]");
    let scrollY = window.scrollY;

    sections.forEach(current => {
        const sectionHeight = current.offsetHeight;
        const sectionTop = current.getBoundingClientRect().top + scrollY - 100;
        const sectionId = current.getAttribute("id");

        // Top nav links
        var navLink = document.querySelector('.nav-link[href*="' + sectionId + '"]');
        if (navLink) {
            if (scrollY > sectionTop && scrollY <= sectionTop + sectionHeight) {
                navLink.classList.add("active");
            } else {
                navLink.classList.remove("active");
            }
        }

        // Sidebar TOC links
        var tocLink = document.querySelector('.toc-link[href="#' + sectionId + '"]');
        if (tocLink) {
            if (scrollY > sectionTop && scrollY <= sectionTop + sectionHeight) {
                tocLink.classList.add("active");
            } else {
                tocLink.classList.remove("active");
            }
        }
    });
}

// === Pointer Tracking ===
(function initPointerTracking() {
    var style = document.createElement('style');
    document.head.appendChild(style);
    var sheet = style.sheet;
    sheet.insertRule('.pointer-relative { --pointer-fixed: 0px 0px; }', 0);
    var rules = sheet.cssRules || sheet.rules;
    var rule = rules[0];

    window.addEventListener("pointermove", function (e) {
        rule.style.setProperty("--pointer-fixed", e.clientX.toFixed(2) + "px " + e.clientY.toFixed(2) + "px");
    }, { passive: true });

    document.addEventListener("pointerleave", function () {
        document.querySelectorAll(".pointer-relative").forEach(function (el) {
            el.classList.add("pointer-none");
        });
    }, { passive: true });

    document.addEventListener("pointerenter", function () {
        document.querySelectorAll(".pointer-relative").forEach(function (el) {
            el.classList.remove("pointer-none");
        });
    }, { passive: true });
})();

// === Project Filter Logic ===
function initFilters() {
    var filterBtns = document.querySelectorAll('.filter-btn');
    var cards = document.querySelectorAll('.cover-date[data-filters]');

    if (!filterBtns.length) return;

    function applyFilter(filterValue) {
        cards.forEach(function(card) {
            var filters = card.getAttribute('data-filters').split(',');
            if (filterValue === 'all') {
                card.classList.remove('filter-hidden');
            } else {
                if (filters.indexOf(filterValue) !== -1) {
                    card.classList.remove('filter-hidden');
                } else {
                    card.classList.add('filter-hidden');
                }
            }
        });
    }

    filterBtns.forEach(function(btn) {
        btn.addEventListener('click', function() {
            filterBtns.forEach(function(b) { b.classList.remove('active'); });
            btn.classList.add('active');
            applyFilter(btn.getAttribute('data-filter'));
        });
    });

    // Apply default filter (all)
    applyFilter('all');
}

// === Sidebar Nav Highlighting ===
function sidebarNavHighlighter() {
    var sections = document.querySelectorAll('.portfolio-content section[id]');
    var navItems = document.querySelectorAll('.sidebar-nav-item');
    var scrollY = window.scrollY;

    sections.forEach(function(section) {
        var sectionHeight = section.offsetHeight;
        var sectionTop = section.getBoundingClientRect().top + scrollY - 120;
        var sectionId = section.getAttribute('id');

        navItems.forEach(function(item) {
            if (item.getAttribute('data-section') === sectionId) {
                if (scrollY > sectionTop && scrollY <= sectionTop + sectionHeight) {
                    item.classList.add('active');
                } else {
                    item.classList.remove('active');
                }
            }
        });
    });
}

// === Initialize ===
updateParallaxLayerSize();
updateRelativeElements();
navHighlighter();
initFilters();

window.addEventListener("scroll", function () {
    updateRelativeElements();
    navHighlighter();
    sidebarNavHighlighter();
}, { passive: true });

window.addEventListener("resize", function () {
    updateParallaxLayerSize();
    updateRelativeElements();
    navHighlighter();
}, { passive: true });
