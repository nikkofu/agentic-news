// js/app.js - The Digital Sanctuary Router

const routes = {
    '': { path: 'pages/home.html', layout: 'app' },
    '#': { path: 'pages/home.html', layout: 'app' },
    '#/home': { path: 'pages/home.html', layout: 'app' },
    '#/article': { path: 'pages/article.html', layout: 'app' },
    '#/category': { path: 'pages/category.html', layout: 'app' },
    '#/community': { path: 'pages/community.html', layout: 'app' },
    '#/knowledge-gap': { path: 'pages/knowledge-gap.html', layout: 'app' },
    '#/learning': { path: 'pages/learning.html', layout: 'app' },
    '#/profile': { path: 'pages/profile.html', layout: 'app' },
    '#/reflection': { path: 'pages/reflection.html', layout: 'app' },
    '#/rss-source': { path: 'pages/rss-source.html', layout: 'app' },
    '#/upgrade': { path: 'pages/upgrade.html', layout: 'app' },
    // Phase 2
    '#/landing': { path: 'pages/landing.html', layout: 'standalone' },
    '#/login': { path: 'pages/login.html', layout: 'standalone' },
    '#/share': { path: 'pages/share.html', layout: 'standalone' },
    '#/invite': { path: 'pages/invite.html', layout: 'app' },
    '#/digest': { path: 'pages/digest.html', layout: 'app' },
    '#/public-profile': { path: 'pages/public-profile.html', layout: 'standalone' },
    '#/analytics': { path: 'pages/analytics.html', layout: 'app' },
    '#/vault': { path: 'pages/vault.html', layout: 'app' },
    '#/briefing': { path: 'pages/briefing.html', layout: 'app' },
    // Phase 3
    '#/onboarding': { path: 'pages/onboarding.html', layout: 'standalone' },
    '#/checkout': { path: 'pages/checkout.html', layout: 'app' },
    '#/settings': { path: 'pages/settings.html', layout: 'app' },
    '#/notifications': { path: 'pages/notifications.html', layout: 'app' },
    '#/404': { path: 'pages/404.html', layout: 'standalone' }
};

const appState = {
    currentPath: ''
};

async function loadPage(hash) {
    const mainContainer = document.getElementById('app-main');
    const routeConfig = routes[hash] || routes['#/home'];
    const path = routeConfig.path;
    const isStandalone = routeConfig.layout === 'standalone';
    
    // UI Feedback
    mainContainer.style.opacity = '0';
    appState.currentPath = hash || '#/home';
    
    // Layout Controller
    const appHeader = document.querySelector('.app-header');
    const bottomNav = document.querySelector('.bottom-nav');
    if (isStandalone) {
        if (appHeader) appHeader.style.display = 'none';
        if (bottomNav) bottomNav.style.display = 'none';
        document.body.classList.remove('pb-32');
    } else {
        if (appHeader) appHeader.style.display = 'block';
        if (bottomNav) bottomNav.style.display = 'flex';
        document.body.classList.add('pb-32');
    }
    
    try {
        const response = await fetch(path);
        if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
        const html = await response.text();
        
        setTimeout(() => {
            mainContainer.innerHTML = html;
            mainContainer.style.opacity = '1';
            updateActiveNav(appState.currentPath);
            window.scrollTo({ top: 0, behavior: 'instant' });
            
            // Re-initialize any dynamic components in the loaded HTML
            initializeComponents();
        }, 150); // slight delay for fade effect

    } catch (e) {
        console.error('Failed to load page:', path, e);
        mainContainer.innerHTML = `
            <div class="container py-20 text-center">
                <span class="material-symbols-outlined text-4xl text-error mb-4">error</span>
                <h2 class="font-headline text-2xl text-primary">Content Unavailable</h2>
                <p class="text-on-surface-variant mt-2">The sanctuary archives are currently restructuring.</p>
                <a href="#/home" class="btn btn-primary mt-6">Return to Atrium</a>
            </div>
        `;
        mainContainer.style.opacity = '1';
    }
}

function updateActiveNav(hash) {
    // Mobile Nav
    document.querySelectorAll('.bottom-nav .nav-item').forEach(el => {
        el.classList.remove('active');
        if (el.getAttribute('href') === hash) {
            el.classList.add('active');
        }
    });

    // Desktop Nav
    document.querySelectorAll('.desktop-nav-link').forEach(el => {
        el.classList.remove('active');
        // Simple matching logic
        if (hash === '#/home' && el.textContent.trim() === 'Briefing' ||
            hash === '#/category' && el.textContent.trim() === 'Domains' ||
            hash === '#/learning' && el.textContent.trim() === 'Learning' ||
            hash === '#/reflection' && el.textContent.trim() === 'Sanctuary'
        ) {
            el.classList.add('active');
        }
    });
}

function initializeComponents() {
    if (appState.currentPath === '#/reflection') {
        hydrateReflectionPage();
        return;
    }
    if (appState.currentPath === '#/community') {
        hydrateCommunityPage();
        return;
    }
    if (appState.currentPath === '#/upgrade') {
        hydrateUpgradePage();
    }
}

async function fetchJSON(url, options) {
    const response = await fetch(url, options);
    if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
    }
    return response.json();
}

function unwrapData(payload) {
    if (payload && typeof payload === 'object' && payload.data !== undefined) {
        return payload.data;
    }
    return payload;
}

function normalizeReflectionCollection(payload) {
    const data = unwrapData(payload);
    if (Array.isArray(data)) return data;
    if (data && Array.isArray(data.items)) return data.items;
    if (data && Array.isArray(data.reflections)) return data.reflections;
    return [];
}

function formatDateTime(value) {
    if (!value) return '';
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) return value;
    return date.toLocaleString();
}

function escapeHTML(value) {
    if (value === undefined || value === null) return '';
    return String(value)
        .replaceAll('&', '&amp;')
        .replaceAll('<', '&lt;')
        .replaceAll('>', '&gt;')
        .replaceAll('"', '&quot;')
        .replaceAll("'", '&#39;');
}

function renderReflectionsList(items, listContainer) {
    if (!Array.isArray(items) || items.length === 0) {
        listContainer.innerHTML = `
            <div class="bg-surface-container-low rounded-2xl border border-surface-container-highest p-4">
                <p class="text-sm text-on-surface-variant">No reflections yet. Your next insight starts above.</p>
            </div>
        `;
        return;
    }

    const sortedItems = [...items].sort((left, right) => {
        const leftTime = Date.parse(left.created_at || '');
        const rightTime = Date.parse(right.created_at || '');
        if (Number.isNaN(leftTime) && Number.isNaN(rightTime)) return 0;
        if (Number.isNaN(leftTime)) return 1;
        if (Number.isNaN(rightTime)) return -1;
        return rightTime - leftTime;
    });

    listContainer.innerHTML = sortedItems.map((item) => {
        const safeContent = escapeHTML(item.content);
        const safeSummary = escapeHTML(item.summary);
        const safeStatus = escapeHTML(item.enhancement_status);
        const tags = Array.isArray(item.tags) ? item.tags : [];
        const tagMarkup = tags.map((tag) => `<span class="px-2 py-1 rounded-full bg-primary/10 text-primary text-[10px] font-bold uppercase tracking-widest">${escapeHTML(tag)}</span>`).join('');
        const createdAt = formatDateTime(item.created_at);

        return `
            <article class="bg-surface-container-low rounded-2xl border border-surface-container-highest p-5 space-y-3">
                ${safeContent ? `<p class="font-body text-sm text-on-surface leading-relaxed">${safeContent}</p>` : ''}
                ${safeSummary ? `<p class="font-body text-xs text-on-surface-variant italic">${safeSummary}</p>` : ''}
                ${(tagMarkup || createdAt || safeStatus) ? `
                    <div class="flex flex-wrap items-center gap-2 pt-1">
                        ${tagMarkup}
                        ${createdAt ? `<span class="text-[10px] font-label uppercase tracking-widest text-secondary">${escapeHTML(createdAt)}</span>` : ''}
                        ${safeStatus ? `<span class="text-[10px] font-label uppercase tracking-widest text-tertiary">${safeStatus}</span>` : ''}
                    </div>
                ` : ''}
            </article>
        `;
    }).join('');
}

async function hydrateReflectionPage() {
    const composeForm = document.querySelector('[data-reflection-compose-form]');
    const reflectionInput = document.querySelector('[data-reflection-input]');
    const listContainer = document.querySelector('[data-reflection-list]');
    const feedback = document.querySelector('[data-reflection-feedback]');

    if (!composeForm || !reflectionInput || !listContainer) return;
    if (composeForm.dataset.hydrated === 'true') return;
    composeForm.dataset.hydrated = 'true';

    const setFeedback = (text, level = 'muted') => {
        if (!feedback) return;
        feedback.textContent = text || '';
        feedback.classList.remove('text-error', 'text-tertiary', 'text-on-surface-variant');
        if (level === 'error') {
            feedback.classList.add('text-error');
            return;
        }
        if (level === 'success') {
            feedback.classList.add('text-tertiary');
            return;
        }
        feedback.classList.add('text-on-surface-variant');
    };

    const loadReflections = async () => {
        try {
            setFeedback('Loading reflections...');
            const payload = await fetchJSON('/api/v1/reflections');
            const items = normalizeReflectionCollection(payload);
            renderReflectionsList(items, listContainer);
            setFeedback('');
        } catch (error) {
            console.error('Failed to load reflections', error);
            setFeedback('Unable to load reflections right now.', 'error');
        }
    };

    composeForm.addEventListener('submit', async (event) => {
        event.preventDefault();
        const content = reflectionInput.value.trim();
        if (!content) {
            setFeedback('Write a reflection before submitting.', 'error');
            return;
        }

        const submitButton = composeForm.querySelector('[data-reflection-submit]');
        if (submitButton) submitButton.disabled = true;
        setFeedback('Sending reflection...');

        try {
            await fetchJSON('/api/v1/reflections', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ content })
            });
            reflectionInput.value = '';
            setFeedback('Reflection stored.', 'success');
            await loadReflections();
        } catch (error) {
            console.error('Failed to submit reflection', error);
            setFeedback('Unable to submit reflection right now.', 'error');
        } finally {
            if (submitButton) submitButton.disabled = false;
        }
    });

    await loadReflections();
}

async function hydrateCommunityPage() {
    const container = document.querySelector('[data-community-preview]');
    if (!container || container.dataset.hydrated === 'true') return;
    container.dataset.hydrated = 'true';

    const headlineNode = container.querySelector('[data-community-headline]');
    const bodyNode = container.querySelector('[data-community-body]');
    const ctaNode = container.querySelector('[data-community-cta]');
    const statusNode = container.querySelector('[data-community-status]');
    const generatedAtNode = container.querySelector('[data-community-generated-at]');

    try {
        const payload = await fetchJSON('/api/v1/community/preview');
        const data = unwrapData(payload) || {};
        if (headlineNode && data.headline) headlineNode.textContent = data.headline;
        if (bodyNode && data.body) bodyNode.textContent = data.body;
        if (ctaNode && data.cta) ctaNode.textContent = data.cta;
        if (statusNode && data.status) statusNode.textContent = data.status;
        if (generatedAtNode && data.generated_at) generatedAtNode.textContent = formatDateTime(data.generated_at);
    } catch (error) {
        console.error('Failed to load community preview', error);
        if (statusNode) statusNode.textContent = 'temporarily unavailable';
    }
}

async function hydrateUpgradePage() {
    const container = document.querySelector('[data-upgrade-offer]');
    if (!container || container.dataset.hydrated === 'true') return;
    container.dataset.hydrated = 'true';

    const headlineNode = document.querySelector('[data-upgrade-headline]');
    const bodyNode = document.querySelector('[data-upgrade-body]');
    const priceNode = container.querySelector('[data-upgrade-price]');
    const statusNode = container.querySelector('[data-upgrade-status]');
    const generatedAtNode = container.querySelector('[data-upgrade-generated-at]');
    const itemsNode = container.querySelector('[data-upgrade-items]');

    try {
        const payload = await fetchJSON('/api/v1/upgrade/offer');
        const data = unwrapData(payload) || {};
        if (headlineNode && data.headline) headlineNode.textContent = data.headline;
        if (bodyNode && data.body) bodyNode.textContent = data.body;
        if (priceNode && data.price_display) priceNode.textContent = data.price_display;
        if (statusNode && data.status) statusNode.textContent = data.status;
        if (generatedAtNode && data.generated_at) generatedAtNode.textContent = formatDateTime(data.generated_at);
        if (itemsNode && Array.isArray(data.offer_items) && data.offer_items.length > 0) {
            itemsNode.innerHTML = data.offer_items
                .map((item) => `<li class="text-xs text-on-surface-variant font-body">${escapeHTML(item)}</li>`)
                .join('');
        }
    } catch (error) {
        console.error('Failed to load upgrade offer', error);
        if (statusNode) statusNode.textContent = 'temporarily unavailable';
    }
}

// Router initialization
window.addEventListener('hashchange', () => {
    loadPage(window.location.hash);
});

// Initial Load
document.addEventListener('DOMContentLoaded', () => {
    loadPage(window.location.hash);
});
