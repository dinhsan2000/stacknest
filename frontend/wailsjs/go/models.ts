export namespace config {
	
	export class ServiceConfig {
	    enabled: boolean;
	    port: number;
	    path: string;
	    version: string;
	
	    static createFrom(source: any = {}) {
	        return new ServiceConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.enabled = source["enabled"];
	        this.port = source["port"];
	        this.path = source["path"];
	        this.version = source["version"];
	    }
	}
	export class Config {
	    root_path: string;
	    bin_path: string;
	    data_path: string;
	    www_path: string;
	    log_path: string;
	    apache: ServiceConfig;
	    nginx: ServiceConfig;
	    mysql: ServiceConfig;
	    php: ServiceConfig;
	    redis: ServiceConfig;
	    auto_start: boolean;
	    theme: string;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.root_path = source["root_path"];
	        this.bin_path = source["bin_path"];
	        this.data_path = source["data_path"];
	        this.www_path = source["www_path"];
	        this.log_path = source["log_path"];
	        this.apache = this.convertValues(source["apache"], ServiceConfig);
	        this.nginx = this.convertValues(source["nginx"], ServiceConfig);
	        this.mysql = this.convertValues(source["mysql"], ServiceConfig);
	        this.php = this.convertValues(source["php"], ServiceConfig);
	        this.redis = this.convertValues(source["redis"], ServiceConfig);
	        this.auto_start = source["auto_start"];
	        this.theme = source["theme"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace configeditor {
	
	export class BackupInfo {
	    path: string;
	    created_at: string;
	    size_bytes: number;
	
	    static createFrom(source: any = {}) {
	        return new BackupInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.created_at = source["created_at"];
	        this.size_bytes = source["size_bytes"];
	    }
	}
	export class ConfigFile {
	    service: string;
	    label: string;
	    path: string;
	    lang: string;
	    writable: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ConfigFile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.service = source["service"];
	        this.label = source["label"];
	        this.path = source["path"];
	        this.lang = source["lang"];
	        this.writable = source["writable"];
	    }
	}

}

export namespace downloader {
	
	export class VersionStatus {
	    version: string;
	    installed: boolean;
	    active: boolean;
	    exe_path: string;
	
	    static createFrom(source: any = {}) {
	        return new VersionStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.installed = source["installed"];
	        this.active = source["active"];
	        this.exe_path = source["exe_path"];
	    }
	}
	export class ServiceVersionStatus {
	    service: string;
	    versions: VersionStatus[];
	
	    static createFrom(source: any = {}) {
	        return new ServiceVersionStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.service = source["service"];
	        this.versions = this.convertValues(source["versions"], VersionStatus);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace logs {
	
	export class LogEntry {
	    service: string;
	    line: string;
	    level: string;
	    timestamp: string;
	
	    static createFrom(source: any = {}) {
	        return new LogEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.service = source["service"];
	        this.line = source["line"];
	        this.level = source["level"];
	        this.timestamp = source["timestamp"];
	    }
	}

}

export namespace phpswitch {
	
	export class PHPInstall {
	    version: string;
	    major: string;
	    path: string;
	    active: boolean;
	
	    static createFrom(source: any = {}) {
	        return new PHPInstall(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.major = source["major"];
	        this.path = source["path"];
	        this.active = source["active"];
	    }
	}

}

export namespace portcheck {
	
	export class ConflictInfo {
	    port: number;
	    pid: number;
	    process: string;
	    in_use: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ConflictInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.port = source["port"];
	        this.pid = source["pid"];
	        this.process = source["process"];
	        this.in_use = source["in_use"];
	    }
	}

}

export namespace project {
	
	export class Project {
	    id: string;
	    name: string;
	    doc_root: string;
	    domain: string;
	    server: string;
	    ssl: boolean;
	    php_path: string;
	    services: Record<string, boolean>;
	    created_at: string;
	    active: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Project(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.doc_root = source["doc_root"];
	        this.domain = source["domain"];
	        this.server = source["server"];
	        this.ssl = source["ssl"];
	        this.php_path = source["php_path"];
	        this.services = source["services"];
	        this.created_at = source["created_at"];
	        this.active = source["active"];
	    }
	}

}

export namespace services {
	
	export class ServiceInfo {
	    name: string;
	    display: string;
	    status: string;
	    port: number;
	    version: string;
	    pid: number;
	    error?: string;
	    enabled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ServiceInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.display = source["display"];
	        this.status = source["status"];
	        this.port = source["port"];
	        this.version = source["version"];
	        this.pid = source["pid"];
	        this.error = source["error"];
	        this.enabled = source["enabled"];
	    }
	}

}

export namespace ssl {
	
	export class CertInfo {
	    domain: string;
	    cert_path: string;
	    key_path: string;
	    expires_at: string;
	
	    static createFrom(source: any = {}) {
	        return new CertInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.domain = source["domain"];
	        this.cert_path = source["cert_path"];
	        this.key_path = source["key_path"];
	        this.expires_at = source["expires_at"];
	    }
	}

}

export namespace vhost {
	
	export class VirtualHost {
	    name: string;
	    domain: string;
	    root: string;
	    ssl: boolean;
	    active: boolean;
	    server: string;
	
	    static createFrom(source: any = {}) {
	        return new VirtualHost(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.domain = source["domain"];
	        this.root = source["root"];
	        this.ssl = source["ssl"];
	        this.active = source["active"];
	        this.server = source["server"];
	    }
	}

}

