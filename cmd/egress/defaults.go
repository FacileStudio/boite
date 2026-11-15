package egress

// DefaultHosts returns the built-in list every allowlist box gets: package
// registries and source hosts needed to install dependencies. Entries are
// exact hosts on ports 80 and 443. Write access to some of these (a GitHub
// repository or gist) can still carry data out; see the README.
func DefaultHosts() []string {
	return []string{
		"deb.debian.org",
		"security.debian.org",
		"cdn-fastly.deb.debian.org",
		"dl-cdn.alpinelinux.org",
		"cache.nixos.org",
		"channels.nixos.org",
		"registry.npmjs.org",
		"registry.yarnpkg.com",
		"pypi.org",
		"files.pythonhosted.org",
		"crates.io",
		"index.crates.io",
		"static.crates.io",
		"proxy.golang.org",
		"sum.golang.org",
		"github.com",
		"api.github.com",
		"codeload.github.com",
		"objects.githubusercontent.com",
		"raw.githubusercontent.com",
		"release-assets.githubusercontent.com",
	}
}
