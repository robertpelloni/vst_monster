CREATE TYPE license_type AS ENUM ('free', 'opensource', 'commercial', 'donationware');
CREATE TYPE install_strategy AS ENUM ('extract_binaries', 'msi_silent', 'exe_silent', 'dmg_extract', 'pkg_silent', 'external_manager');

CREATE TABLE plugins (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    developer VARCHAR(255) NOT NULL,
    license_model license_type NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(name, developer)
);

CREATE TABLE plugin_releases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    plugin_id UUID REFERENCES plugins(id) ON DELETE CASCADE,
    version VARCHAR(50) NOT NULL,
    release_date TIMESTAMP WITH TIME ZONE,
    UNIQUE(plugin_id, version)
);

CREATE TABLE plugin_distributions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    release_id UUID REFERENCES plugin_releases(id) ON DELETE CASCADE,
    platform VARCHAR(50) NOT NULL, -- 'windows', 'macos', 'linux'
    architecture VARCHAR(50) NOT NULL, -- 'x86_64', 'arm64'
    download_url TEXT NOT NULL,
    sha256_hash VARCHAR(64) NOT NULL,
    strategy install_strategy NOT NULL,
    extraction_rules JSONB NOT NULL, -- e.g., {"target_binaries": ["Plugin.vst3"], "sub_path": "/bin"}
    is_active BOOLEAN DEFAULT TRUE
);
