export type Locale = 'en' | 'vi'

export interface Translations {
    // Common
    loading: string
    save: string
    saving: string
    cancel: string
    delete: string
    start: string
    stop: string
    restart: string
    open: string
    close: string
    dismiss: string
    yes: string
    no: string
    error: string
    success: string
    confirm: string
    browse: string

    // Sidebar
    nav_dashboard: string
    nav_binaries: string
    nav_database: string
    nav_vhosts: string
    nav_config: string
    nav_php: string
    nav_logs: string
    nav_terminal: string
    nav_settings: string

    // Header
    header_services_running: string
    header_all_stopped: string
    header_minimize: string

    // Dashboard
    dash_title: string
    dash_running_count: string
    dash_stop_all: string
    dash_start_all: string
    dash_service: string
    dash_status: string
    dash_port: string
    dash_pid: string
    dash_on: string
    dash_actions: string
    dash_loading: string
    dash_starting: string
    dash_enabled_tip: string
    dash_disabled_tip: string
    dash_version: string
    dash_switch_version: string
    dash_switching: string

    // Service status
    status_running: string
    status_stopped: string
    status_starting: string
    status_stopping: string
    status_error: string

    // Binaries
    bin_title: string
    bin_desc: string
    bin_missing_warn: string
    bin_not_installed: string
    bin_active: string
    bin_installed: string
    bin_set_active: string
    bin_download: string
    bin_downloading: string
    bin_connecting: string
    bin_cancel: string
    bin_confirm_delete: string
    bin_footer_1: string
    bin_footer_2: string
    bin_footer_3: string

    // Virtual Hosts
    vh_title: string
    vh_ca_title: string
    vh_ca_trusted: string
    vh_ca_not_trusted: string
    vh_trust_ca: string
    vh_trusting: string
    vh_export_ca: string
    vh_add_title: string
    vh_name_placeholder: string
    vh_domain_placeholder: string
    vh_root_placeholder: string
    vh_enable_ssl: string
    vh_add_host: string
    vh_no_hosts: string
    vh_remove: string
    vh_remove_confirm: string
    vh_yes_remove: string
    vh_no_cert: string
    vh_cert_expires: string
    vh_regenerate: string
    vh_generate_cert: string
    vh_generating: string

    // Config Editor
    cfg_title: string
    cfg_desc: string
    cfg_files: string
    cfg_readonly: string
    cfg_unsaved: string
    cfg_backups: string
    cfg_hide_backups: string
    cfg_save_restart: string
    cfg_restarting: string
    cfg_saved_ok: string
    cfg_no_files: string
    cfg_no_backups: string
    cfg_select_file: string
    cfg_available_backups: string
    cfg_restore: string
    cfg_restore_confirm: string
    cfg_unsaved_discard: string

    // Log Viewer
    log_title: string
    log_lines: string
    log_errors: string
    log_warnings: string
    log_clear: string
    log_pause: string
    log_resume: string
    log_autoscroll: string
    log_filter: string
    log_all_levels: string
    log_no_entries: string
    log_start_to_see: string
    log_watching: string
    log_lines_shown: string

    // Terminal
    term_title: string
    term_running: string
    term_exited: string
    term_not_started: string
    term_open_www: string
    term_start: string
    term_start_btn: string
    term_open_www_btn: string
    term_click_start: string

    // Database
    db_title: string
    db_desc: string
    db_mysql_start_warn: string
    db_adminer: string
    db_adminer_desc: string
    db_web_based: string
    db_native_client: string
    db_start_open: string
    db_starting: string
    db_adminer_not_found: string
    db_php_not_found: string
    db_heidisql: string
    db_heidisql_desc: string
    db_open_heidisql: string
    db_credentials_title: string
    db_server: string
    db_username: string
    db_password: string
    db_password_empty: string

    // PHP Switcher
    php_title: string
    php_active: string
    php_no_found: string
    php_add_path: string
    php_rescan: string
    php_scanning: string
    php_no_php_title: string
    php_no_php_desc: string
    php_add_custom: string
    php_in_use: string
    php_use_version: string
    php_switching: string
    php_how_title: string
    php_how_1: string
    php_how_2: string
    php_how_3: string
    php_how_4: string

    // Settings
    settings_title: string
    settings_paths: string
    settings_root: string
    settings_data: string
    settings_data_hint: string
    settings_www: string
    settings_log: string
    settings_ports: string
    settings_port_range: string
    settings_port_conflict: string
    settings_general: string
    settings_autostart: string
    settings_saved: string
    settings_fix_errors: string
    settings_language: string

    // Port Conflict Modal
    pc_title: string
    pc_port_in_use: string
    pc_unable_start: string
    pc_process: string
    pc_killing: string
    pc_kill_start: string
    pc_free_port: string
}
