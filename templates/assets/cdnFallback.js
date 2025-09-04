// cdnFallback.js
(function (global) {
    // 默认超时时间，3秒
    const DEFAULT_TIMEOUT = 3000;

    // 配置文件
    const config = {
        timeout: DEFAULT_TIMEOUT, // 超时时间
        cssFallbackDir: '/assets', // 本地 CSS 目录
        jsFallbackDir: '/assets'  // 本地 JS 目录
    };

    // 加载 JS 文件
    function loadJS(cdnUrl, localFile = config.jsFallbackDir, timeout = config.timeout) {
        return new Promise((resolve, reject) => {
            const script = document.createElement('script');
            let done = false;

            script.src = cdnUrl;
            script.async = true;

            const timer = setTimeout(() => {
                if (!done) {
                    console.warn(`CDN JS (${cdnUrl}) timeout, switching to local.`);
                    document.head.removeChild(script);
                    loadLocalJS(localFile);
                }
            }, timeout);

            script.onload = () => {
                done = true;
                clearTimeout(timer);
                resolve();
            };

            script.onerror = () => {
                done = true;
                clearTimeout(timer);
                loadLocalJS(localFile);
            };

            function loadLocalJS(localUrl) {
                const localScript = document.createElement('script');
                localScript.src = localUrl;
                localScript.onload = resolve;
                localScript.onerror = reject;
                document.head.appendChild(localScript);
            }

            document.head.appendChild(script);
        });
    }

    // 加载 CSS 文件
    function loadCSS(cdnUrl, localFile = config.cssFallbackDir, timeout = config.timeout) {
        return new Promise((resolve, reject) => {
            const link = document.createElement('link');
            let done = false;

            link.rel = 'stylesheet';
            link.href = cdnUrl;

            const timer = setTimeout(() => {
                if (!done) {
                    console.warn(`CDN CSS (${cdnUrl}) timeout, switching to local.`);
                    document.head.removeChild(link);
                    loadLocalCSS(localFile);
                }
            }, timeout);

            link.onload = () => {
                done = true;
                clearTimeout(timer);
                resolve();
            };

            link.onerror = () => {
                done = true;
                clearTimeout(timer);
                loadLocalCSS(localFile);
            };

            function loadLocalCSS(localUrl) {
                const localLink = document.createElement('link');
                localLink.rel = 'stylesheet';
                localLink.href = localUrl;
                localLink.onload = resolve;
                localLink.onerror = reject;
                document.head.appendChild(localLink);
            }

            document.head.appendChild(link);
        });
    }

    // 提供外部接口
    global.loadCDN = {
        loadJS: loadJS,
        loadCSS: loadCSS,
        setConfig: (newConfig) => {
            Object.assign(config, newConfig);
        }
    };
})(window);
