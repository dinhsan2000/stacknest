<?php
// ─── Runtime info ─────────────────────────────────────────────────────────────
$phpVersion   = phpversion();
$phpBuild     = PHP_SAPI;
$phpOs        = PHP_OS;

// MySQL
$mysqlVersion = null;
$mysqlError   = null;
try {
    $pdo = new PDO('mysql:host=127.0.0.1;port=3306', 'root', '', [
        PDO::ATTR_TIMEOUT    => 2,
        PDO::ATTR_ERRMODE    => PDO::ERRMODE_EXCEPTION,
    ]);
    $mysqlVersion = $pdo->query('SELECT VERSION()')->fetchColumn();
} catch (Throwable $e) {
    $mysqlError = $e->getMessage();
}

// Redis
$redisOk    = false;
$redisError = null;
if (class_exists('Redis')) {
    try {
        $r = new Redis();
        $r->connect('127.0.0.1', 6379, 1);
        $redisOk = ($r->ping() === true || $r->ping() === '+PONG');
    } catch (Throwable $e) {
        $redisError = $e->getMessage();
    }
} else {
    $redisError = 'ext-redis not loaded';
}

// Web server
$webServer = $_SERVER['SERVER_SOFTWARE'] ?? 'Unknown';
$docRoot   = $_SERVER['DOCUMENT_ROOT'] ?? realpath(__DIR__);
$port      = $_SERVER['SERVER_PORT'] ?? '80';
$host      = gethostname();

// Extensions
$exts = get_loaded_extensions();
sort($exts);

// Helpers
function badge(bool $ok, string $label = ''): string {
    $color = $ok ? '#22c55e' : '#ef4444';
    $text  = $label ?: ($ok ? 'Running' : 'Error');
    return "<span style='display:inline-flex;align-items:center;gap:5px;font-size:12px;font-weight:600;color:{$color}'>
                <span style='width:8px;height:8px;border-radius:50%;background:{$color};display:inline-block'></span>
                {$text}
            </span>";
}
?>
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>StackNest</title>
    <style>
        *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0 }

        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
            background: #0f1420;
            color: #c9d1d9;
            min-height: 100vh;
            padding: 32px 16px 64px;
        }

        /* ── Header ── */
        .header {
            text-align: center;
            margin-bottom: 40px;
        }
        .header-logo {
            display: inline-flex;
            align-items: center;
            gap: 12px;
            margin-bottom: 8px;
        }
        .header-logo svg { width: 40px; height: 40px }
        .header h1 {
            font-size: 28px;
            font-weight: 700;
            color: #e6edf3;
            letter-spacing: -0.5px;
        }
        .header p {
            font-size: 14px;
            color: #6e7681;
            margin-top: 4px;
        }

        /* ── Grid ── */
        .grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
            gap: 16px;
            max-width: 1100px;
            margin: 0 auto 32px;
        }

        /* ── Card ── */
        .card {
            background: #161b27;
            border: 1px solid #21293a;
            border-radius: 12px;
            padding: 20px 22px;
        }
        .card-header {
            display: flex;
            justify-content: space-between;
            align-items: flex-start;
            margin-bottom: 14px;
        }
        .card-icon {
            width: 40px;
            height: 40px;
            border-radius: 10px;
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 20px;
        }
        .card-title {
            font-size: 13px;
            font-weight: 600;
            color: #8b949e;
            text-transform: uppercase;
            letter-spacing: 0.6px;
            margin-bottom: 6px;
        }
        .card-value {
            font-size: 22px;
            font-weight: 700;
            color: #e6edf3;
            letter-spacing: -0.3px;
        }
        .card-sub {
            font-size: 12px;
            color: #6e7681;
            margin-top: 4px;
        }
        .card-meta {
            margin-top: 14px;
            padding-top: 14px;
            border-top: 1px solid #21293a;
            font-size: 12px;
            color: #6e7681;
            line-height: 1.7;
        }
        .card-meta strong { color: #8b949e }

        /* ── Extensions ── */
        .section {
            max-width: 1100px;
            margin: 0 auto 24px;
        }
        .section-title {
            font-size: 13px;
            font-weight: 600;
            color: #8b949e;
            text-transform: uppercase;
            letter-spacing: 0.6px;
            margin-bottom: 12px;
        }
        .ext-grid {
            display: flex;
            flex-wrap: wrap;
            gap: 6px;
        }
        .ext-tag {
            background: #1e2535;
            border: 1px solid #2a3347;
            border-radius: 6px;
            padding: 3px 9px;
            font-size: 12px;
            color: #8b949e;
        }

        /* ── Server table ── */
        .info-table { width: 100%; border-collapse: collapse }
        .info-table td {
            padding: 8px 12px;
            font-size: 13px;
            border-bottom: 1px solid #21293a;
            vertical-align: top;
        }
        .info-table td:first-child {
            color: #8b949e;
            width: 180px;
            font-weight: 500;
        }
        .info-table td:last-child { color: #c9d1d9; word-break: break-all }
        .info-table tr:last-child td { border-bottom: none }

        /* ── Footer ── */
        .footer {
            text-align: center;
            font-size: 12px;
            color: #3d444d;
            margin-top: 48px;
        }
        .footer a { color: #388bfd; text-decoration: none }
    </style>
</head>
<body>

<!-- Header -->
<div class="header">
    <div class="header-logo">
        <svg viewBox="0 0 40 40" fill="none" xmlns="http://www.w3.org/2000/svg">
            <rect width="40" height="40" rx="10" fill="#1e2535"/>
            <path d="M8 28 L20 12 L32 28" stroke="#3b82f6" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" fill="none"/>
            <path d="M12 28 L20 16 L28 28" stroke="#60a5fa" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" fill="none"/>
            <circle cx="20" cy="28" r="2.5" fill="#3b82f6"/>
        </svg>
        <h1>StackNest</h1>
    </div>
    <p>Local Development Environment &mdash; <?= htmlspecialchars($host) ?></p>
</div>

<!-- Service Cards -->
<div class="grid">

    <!-- PHP -->
    <div class="card">
        <div class="card-header">
            <div>
                <div class="card-title">PHP</div>
                <div class="card-value"><?= htmlspecialchars($phpVersion) ?></div>
                <div class="card-sub"><?= htmlspecialchars(PHP_SAPI) ?> &mdash; <?= htmlspecialchars(PHP_OS) ?></div>
            </div>
            <div class="card-icon" style="background:#1e3a2f"><?= badge(true, 'Active') ?></div>
        </div>
        <div class="card-meta">
            <strong>ini:</strong> <?= htmlspecialchars(php_ini_loaded_file() ?: 'none') ?><br>
            <strong>Extensions:</strong> <?= count($exts) ?> loaded<br>
            <strong>Memory limit:</strong> <?= ini_get('memory_limit') ?><br>
            <strong>Max upload:</strong> <?= ini_get('upload_max_filesize') ?>
        </div>
    </div>

    <!-- MySQL -->
    <div class="card">
        <div class="card-header">
            <div>
                <div class="card-title">MySQL</div>
                <div class="card-value"><?= $mysqlVersion ? htmlspecialchars($mysqlVersion) : '—' ?></div>
                <div class="card-sub">127.0.0.1:3306</div>
            </div>
            <div class="card-icon" style="background:<?= $mysqlVersion ? '#1e3a2f' : '#3a1e1e' ?>">
                <?= badge($mysqlVersion !== null, $mysqlVersion ? 'Running' : 'Error') ?>
            </div>
        </div>
        <div class="card-meta">
            <?php if ($mysqlVersion): ?>
                <strong>Host:</strong> 127.0.0.1<br>
                <strong>User:</strong> root<br>
                <strong>Version:</strong> <?= htmlspecialchars($mysqlVersion) ?>
            <?php else: ?>
                <span style="color:#ef4444"><?= htmlspecialchars($mysqlError ?? 'Cannot connect') ?></span>
            <?php endif ?>
        </div>
    </div>

    <!-- Web Server -->
    <div class="card">
        <div class="card-header">
            <div>
                <div class="card-title">Web Server</div>
                <div class="card-value" style="font-size:16px"><?= htmlspecialchars($webServer) ?></div>
                <div class="card-sub">Port <?= htmlspecialchars($port) ?></div>
            </div>
            <div class="card-icon" style="background:#1e3a2f"><?= badge(true, 'Running') ?></div>
        </div>
        <div class="card-meta">
            <strong>Document Root:</strong><br>
            <span style="word-break:break-all"><?= htmlspecialchars($docRoot) ?></span>
        </div>
    </div>

    <!-- Redis -->
    <div class="card">
        <div class="card-header">
            <div>
                <div class="card-title">Redis</div>
                <div class="card-value"><?= $redisOk ? '127.0.0.1:6379' : '—' ?></div>
                <div class="card-sub"><?= class_exists('Redis') ? 'ext-redis loaded' : 'ext-redis not available' ?></div>
            </div>
            <div class="card-icon" style="background:<?= $redisOk ? '#1e3a2f' : '#2a2a1e' ?>">
                <?= badge($redisOk, $redisOk ? 'Running' : 'Stopped') ?>
            </div>
        </div>
        <div class="card-meta">
            <?php if ($redisError): ?>
                <span style="color:#8b949e"><?= htmlspecialchars($redisError) ?></span>
            <?php else: ?>
                <strong>Host:</strong> 127.0.0.1<br>
                <strong>Port:</strong> 6379
            <?php endif ?>
        </div>
    </div>

</div>

<!-- PHP Extensions -->
<div class="section">
    <div class="card">
        <div class="section-title" style="margin-bottom:12px">Loaded Extensions (<?= count($exts) ?>)</div>
        <div class="ext-grid">
            <?php foreach ($exts as $ext): ?>
                <span class="ext-tag"><?= htmlspecialchars($ext) ?></span>
            <?php endforeach ?>
        </div>
    </div>
</div>

<!-- Server Environment -->
<div class="section">
    <div class="card">
        <div class="section-title" style="margin-bottom:12px">Server Environment</div>
        <table class="info-table">
            <tr><td>PHP Version</td><td><?= htmlspecialchars(phpversion()) ?></td></tr>
            <tr><td>Zend Engine</td><td><?= htmlspecialchars(zend_version()) ?></td></tr>
            <tr><td>SAPI</td><td><?= htmlspecialchars(PHP_SAPI) ?></td></tr>
            <tr><td>OS</td><td><?= htmlspecialchars(PHP_OS_FAMILY . ' ' . php_uname('r')) ?></td></tr>
            <tr><td>Architecture</td><td><?= PHP_INT_SIZE === 8 ? '64-bit' : '32-bit' ?></td></tr>
            <tr><td>Server Software</td><td><?= htmlspecialchars($webServer) ?></td></tr>
            <tr><td>Document Root</td><td><?= htmlspecialchars($docRoot) ?></td></tr>
            <tr><td>Request Time</td><td><?= date('Y-m-d H:i:s') ?></td></tr>
            <tr><td>Timezone</td><td><?= htmlspecialchars(date_default_timezone_get()) ?></td></tr>
            <tr><td>max_execution_time</td><td><?= ini_get('max_execution_time') ?>s</td></tr>
            <tr><td>post_max_size</td><td><?= ini_get('post_max_size') ?></td></tr>
            <tr><td>upload_max_filesize</td><td><?= ini_get('upload_max_filesize') ?></td></tr>
        </table>
    </div>
</div>

<div class="footer">
    Powered by <a href="https://github.com" target="_blank">StackNest</a>
    &mdash; <?= date('Y') ?>
</div>

</body>
</html>
