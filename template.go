// SPDX-License-Identifier: ISC

package main

import (
	"crypto/sha256"
	"html/template"
)

var (
	indexTmpl    = template.Must(template.New("index").Parse(htmlTemplate))
	templateHash = sha256.Sum256([]byte(htmlTemplate))
)

const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta name="color-scheme" content="light dark">
    <title>chute</title>
    <script>
        (function() {
            const saved = localStorage.getItem('chute_theme') || 'auto';
            if (saved === 'dark') {
                document.documentElement.setAttribute('data-theme', 'dark');
                document.documentElement.style.colorScheme = 'dark';
            } else if (saved === 'light') {
                document.documentElement.setAttribute('data-theme', 'light');
                document.documentElement.style.colorScheme = 'light';
            } else {
                document.documentElement.removeAttribute('data-theme');
                document.documentElement.style.colorScheme = 'light dark';
            }
        })();
    </script>
    <style>
        :root,
        [data-theme="light"] {
            --bg-color: #f8fafc;
            --card-bg: #ffffff;
            --card-border: #e2e8f0;
            --text-color: #0f172a;
            --text-muted: #64748b;
            --primary: #2563eb;
            --primary-hover: #1d4ed8;
            --drop-bg: #eff6ff;
            --drop-text: #1d4ed8;
            --drop-dragover-bg: #dbeafe;
            --drop-dragover-border: #1d4ed8;
            --card-shadow: 0 1px 3px rgba(0, 0, 0, 0.08), 0 1px 2px rgba(0, 0, 0, 0.04);
            --item-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
            --toggle-active-bg: #ffffff;
            --toggle-active-shadow: 0 1px 2px rgba(0, 0, 0, 0.12);
        }

        [data-theme="dark"] {
            --bg-color: #0f172a;
            --card-bg: #1e293b;
            --card-border: #334155;
            --text-color: #f1f5f9;
            --text-muted: #94a3b8;
            --primary: #60a5fa;
            --primary-hover: #93c5fd;
            --drop-bg: rgba(59, 130, 246, 0.08);
            --drop-text: #60a5fa;
            --drop-dragover-bg: rgba(59, 130, 246, 0.2);
            --drop-dragover-border: #93c5fd;
            --card-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.3), 0 2px 4px -2px rgba(0, 0, 0, 0.3);
            --item-shadow: 0 1px 3px rgba(0, 0, 0, 0.2);
            --toggle-active-bg: #0f172a;
            --toggle-active-shadow: 0 1px 2px rgba(0, 0, 0, 0.4);
        }

        @media (prefers-color-scheme: dark) {
            :root:not([data-theme="light"]) {
                --bg-color: #0f172a;
                --card-bg: #1e293b;
                --card-border: #334155;
                --text-color: #f1f5f9;
                --text-muted: #94a3b8;
                --primary: #60a5fa;
                --primary-hover: #93c5fd;
                --drop-bg: rgba(59, 130, 246, 0.08);
                --drop-text: #60a5fa;
                --drop-dragover-bg: rgba(59, 130, 246, 0.2);
                --drop-dragover-border: #93c5fd;
                --card-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.3), 0 2px 4px -2px rgba(0, 0, 0, 0.3);
                --item-shadow: 0 1px 3px rgba(0, 0, 0, 0.2);
                --toggle-active-bg: #0f172a;
                --toggle-active-shadow: 0 1px 2px rgba(0, 0, 0, 0.4);
            }
        }

        *, *::before, *::after {
            box-sizing: border-box;
        }

        body {
            font-family: system-ui, sans-serif;
            max-width: 600px;
            margin: 2rem auto;
            padding: 0 1rem;
            line-height: 1.5;
            background-color: var(--bg-color);
            color: var(--text-color);
            transition: background-color 0.2s ease, color 0.2s ease;
        }

        .header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            gap: 1rem;
            margin-bottom: 1.5rem;
            flex-wrap: wrap;
        }

        .header h2 {
            margin: 0;
            font-size: 1.5rem;
            font-weight: 700;
        }

        .theme-toggle {
            display: inline-flex;
            background: var(--card-border);
            border-radius: 20px;
            padding: 3px;
            gap: 2px;
        }

        .theme-btn {
            display: inline-flex;
            align-items: center;
            gap: 5px;
            border: none;
            background: transparent;
            color: var(--text-muted);
            font-size: 0.8rem;
            font-weight: 500;
            padding: 5px 10px;
            border-radius: 16px;
            cursor: pointer;
            transition: all 0.15s ease;
            user-select: none;
            line-height: 1;
        }

        .theme-btn svg {
            flex-shrink: 0;
        }

        .theme-btn:hover {
            color: var(--text-color);
        }

        .theme-btn.active {
            background: var(--toggle-active-bg);
            color: var(--text-color);
            box-shadow: var(--toggle-active-shadow);
        }

        .card {
            background-color: var(--card-bg);
            border: 1px solid var(--card-border);
            padding: 1.5rem;
            border-radius: 10px;
            box-shadow: var(--card-shadow);
            margin-bottom: 1.5rem;
            transition: background-color 0.2s ease, border-color 0.2s ease;
        }

        .drop-zone {
            border: 2px dashed #3b82f6;
            border-radius: 8px;
            padding: 2.5rem 1rem;
            text-align: center;
            background-color: var(--drop-bg);
            cursor: pointer;
            transition: all 0.2s ease;
        }

        .drop-zone.dragover {
            background-color: var(--drop-dragover-bg);
            border-color: var(--drop-dragover-border);
            transform: scale(1.02);
        }

        .drop-zone p {
            margin: 0;
            color: var(--drop-text);
            font-weight: 500;
        }

        .drop-zone small {
            color: var(--text-muted);
            font-size: 0.85rem;
        }

        #fileInput {
            display: none;
        }

        progress {
            width: 100%;
            height: 12px;
            margin-top: 1rem;
            display: none;
            appearance: none;
            border: none;
            border-radius: 6px;
            overflow: hidden;
            background-color: var(--card-border);
            accent-color: var(--primary);
        }

        progress::-webkit-progress-bar {
            background-color: var(--card-border);
            border-radius: 6px;
        }

        progress::-webkit-progress-value {
            background-color: var(--primary);
            border-radius: 6px;
            transition: width 0.1s ease;
        }

        progress::-moz-progress-bar {
            background-color: var(--primary);
            border-radius: 6px;
        }

        #status {
            font-weight: bold;
            margin-top: 0.5rem;
            text-align: center;
            color: var(--text-color);
        }

        h3 {
            color: var(--text-color);
            margin-top: 1.5rem;
            margin-bottom: 0.75rem;
        }

        ul {
            list-style: none;
            padding: 0;
            margin: 0;
        }

        li {
            background-color: var(--card-bg);
            border: 1px solid var(--card-border);
            margin: 0.5rem 0;
            padding: 0.75rem 1rem;
            border-radius: 6px;
            display: flex;
            justify-content: space-between;
            align-items: center;
            box-shadow: var(--item-shadow);
            transition: background-color 0.2s ease, border-color 0.2s ease;
        }

        li span {
            word-break: break-all;
            margin-right: 0.5rem;
        }

        li small {
            color: var(--text-muted);
            font-size: 0.85rem;
        }

        a {
            color: var(--primary);
            text-decoration: none;
            font-weight: 500;
            white-space: nowrap;
        }

        a:hover {
            color: var(--primary-hover);
            text-decoration: underline;
        }
    </style>
</head>
<body>
    <header class="header">
        <h2>chute</h2>
        <div class="theme-toggle" role="group" aria-label="Theme selector">
            <button type="button" class="theme-btn" data-mode="light" aria-pressed="false" title="Light theme">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="5"></circle><line x1="12" y1="1" x2="12" y2="3"></line><line x1="12" y1="21" x2="12" y2="23"></line><line x1="4.22" y1="4.22" x2="5.64" y2="5.64"></line><line x1="18.36" y1="18.36" x2="19.78" y2="19.78"></line><line x1="1" y1="12" x2="3" y2="12"></line><line x1="21" y1="12" x2="23" y2="12"></line><line x1="4.22" y1="19.78" x2="5.64" y2="18.36"></line><line x1="18.36" y1="5.64" x2="19.78" y2="4.22"></line></svg>
                Light
            </button>
            <button type="button" class="theme-btn" data-mode="dark" aria-pressed="false" title="Dark theme">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"></path></svg>
                Dark
            </button>
            <button type="button" class="theme-btn active" data-mode="auto" aria-pressed="true" title="System theme">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="2" y="3" width="20" height="14" rx="2" ry="2"></rect><line x1="8" y1="21" x2="16" y2="21"></line><line x1="12" y1="17" x2="12" y2="21"></line></svg>
                Auto
            </button>
        </div>
    </header>

    <div class="card">
        <div id="dropZone" class="drop-zone">
            <p>Drag & Drop files here</p>
            <small>or tap to browse files</small>
            <input type="file" id="fileInput" name="files" multiple>
        </div>
        <progress id="progressBar" value="0" max="100"></progress>
        <p id="status"></p>
    </div>

    <h3>Uploaded Files</h3>
    <ul>
        {{range .Files}}
            <li>
                <span>{{.Name}} <small>({{.Size}})</small></span>
                <a href="/uploads/{{.Name}}" download>Download</a>
            </li>
        {{else}}
            <li>No files uploaded yet.</li>
        {{end}}
    </ul>

    <script>
        const dropZone = document.getElementById('dropZone');
        const fileInput = document.getElementById('fileInput');
        const progressBar = document.getElementById('progressBar');
        const status = document.getElementById('status');

        function applyTheme(mode) {
            if (mode === 'dark') {
                document.documentElement.setAttribute('data-theme', 'dark');
                document.documentElement.style.colorScheme = 'dark';
            } else if (mode === 'light') {
                document.documentElement.setAttribute('data-theme', 'light');
                document.documentElement.style.colorScheme = 'light';
            } else {
                document.documentElement.removeAttribute('data-theme');
                document.documentElement.style.colorScheme = 'light dark';
            }
            document.querySelectorAll('.theme-btn').forEach(btn => {
                const isActive = btn.dataset.mode === mode;
                btn.classList.toggle('active', isActive);
                btn.setAttribute('aria-pressed', isActive ? 'true' : 'false');
            });
        }

        const savedTheme = localStorage.getItem('chute_theme') || 'auto';
        applyTheme(savedTheme);

        document.querySelectorAll('.theme-btn').forEach(btn => {
            btn.addEventListener('click', () => {
                const mode = btn.dataset.mode;
                localStorage.setItem('chute_theme', mode);
                applyTheme(mode);
            });
        });

        window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', () => {
            const current = localStorage.getItem('chute_theme') || 'auto';
            if (current === 'auto') {
                applyTheme('auto');
            }
        });

        dropZone.addEventListener('click', () => fileInput.click());
        fileInput.addEventListener('change', () => { if (fileInput.files.length) uploadFiles(fileInput.files); });

        ['dragenter', 'dragover'].forEach(evt => {
            dropZone.addEventListener(evt, (e) => { e.preventDefault(); dropZone.classList.add('dragover'); });
        });

        ['dragleave', 'drop'].forEach(evt => {
            dropZone.addEventListener(evt, (e) => { e.preventDefault(); dropZone.classList.remove('dragover'); });
        });

        dropZone.addEventListener('drop', (e) => {
            const files = e.dataTransfer.files;
            if (files.length) uploadFiles(files);
        });

        function formatSpeed(bytesPerSec) {
            if (bytesPerSec >= 1048576) return (bytesPerSec / 1048576).toFixed(1) + ' MB/s';
            if (bytesPerSec >= 1024) return (bytesPerSec / 1024).toFixed(0) + ' KB/s';
            return bytesPerSec.toFixed(0) + ' B/s';
        }

        function formatETA(seconds) {
            if (seconds < 60) return Math.ceil(seconds) + 's remaining';
            const m = Math.floor(seconds / 60);
            const s = Math.ceil(seconds % 60);
            return m + 'm ' + s + 's remaining';
        }

        function uploadFiles(files) {
            const formData = new FormData();
            for (let i = 0; i < files.length; i++) formData.append('files', files[i]);

            const xhr = new XMLHttpRequest();
            const startTime = Date.now();
            progressBar.style.display = 'block';
            status.innerText = 'Uploading...';

            xhr.upload.onprogress = (e) => {
                if (e.lengthComputable) {
                    const percent = Math.round((e.loaded / e.total) * 100);
                    progressBar.value = percent;

                    const elapsed = (Date.now() - startTime) / 1000;
                    if (elapsed >= 0.5 && e.loaded > 0) {
                        const speed = e.loaded / elapsed;
                        const remaining = (e.total - e.loaded) / speed;
                        status.innerText = 'Uploading: ' + percent + '% \u00b7 ' + formatSpeed(speed) + ' \u00b7 ~' + formatETA(remaining);
                    } else {
                        status.innerText = 'Uploading: ' + percent + '%';
                    }
                }
            };

            xhr.onload = () => {
                if (xhr.status === 200) {
                    status.innerText = 'Upload successful!';
                    setTimeout(() => window.location.reload(), 800);
                } else {
                    status.innerText = 'Upload failed!';
                }
            };

            xhr.open('POST', '/upload', true);
            xhr.send(formData);
        }
    </script>
</body>
</html>
`
