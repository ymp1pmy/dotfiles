// dotfiles installer: symlink manager + system package installer.
// Handles Linux/macOS on amd64/arm64 via runtime.GOOS / runtime.GOARCH.
// Usage: ./install  (via the arch-detecting wrapper script)
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"gopkg.in/yaml.v3"
)

func main() {
	configPath := flag.String("c", "install.conf.yaml", "config file")
	flag.Parse()

	absConfig, err := filepath.Abs(*configPath)
	if err != nil {
		die("resolve config: %v", err)
	}
	dotfilesDir := filepath.Dir(absConfig)

	logf("dotfiles install  os=%s  arch=%s", runtime.GOOS, runtime.GOARCH)

	removeLegacyLocalSymlink()
	ensureConfigWritable()
	installPackages()

	cfg, err := parseConfig(absConfig)
	if err != nil {
		die("parse config: %v", err)
	}
	execute(cfg, dotfilesDir)
	linkConfigSubdirs(dotfilesDir)

	writeStarshipLocalEnv(dotfilesDir)
	installMise()
	gitSkipWorktree(dotfilesDir)
	setupWSL(dotfilesDir)

	logf("Done.")
}

// ── pre-install cleanup ────────────────────────────────────────────────────

// removeLegacyLocalSymlink removes ~/.local if it's a legacy symlink so we
// can replace it with selective per-directory symlinks.
func removeLegacyLocalSymlink() {
	home, _ := os.UserHomeDir()
	p := filepath.Join(home, ".local")
	info, err := os.Lstat(p)
	if err != nil || info.Mode()&os.ModeSymlink == 0 {
		return
	}
	logf("removing legacy ~/.local symlink")
	if err := os.Remove(p); err != nil {
		warnf("remove ~/.local: %v", err)
	}
}

// ensureConfigWritable fixes ~/.config ownership when it's root-owned (common
// in devcontainer setups that pre-create the dir before user provisioning runs).
func ensureConfigWritable() {
	if runtime.GOOS != "linux" {
		return
	}
	home, _ := os.UserHomeDir()
	configDir := filepath.Join(home, ".config")
	testFile := filepath.Join(configDir, ".write-test")
	f, err := os.Create(testFile)
	if err == nil {
		f.Close()
		os.Remove(testFile)
		return
	}
	logf("~/.config not writable — fixing ownership with sudo...")
	mustRun("sudo", "chown", fmt.Sprintf("%d", os.Getuid()), configDir)
}

// ── OS/arch package installation ───────────────────────────────────────────

func installPackages() {
	switch runtime.GOOS {
	case "darwin":
		installDarwin()
	case "linux":
		installLinux()
	default:
		warnf("unknown OS %q — skipping package install", runtime.GOOS)
	}
}

func installDarwin() {
	logf("installing macOS packages...")
	brew := brewPath()
	if brew == "" {
		if isTTY() {
			fmt.Print("Homebrew not found. Press Enter to install or Ctrl+C to cancel... ")
			fmt.Scanln()
		} else {
			logf("Homebrew not found — installing automatically...")
		}
		runInteractive("/bin/bash", "-c",
			`curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh | /bin/bash`)
		brew = brewPath()
	}
	if brew == "" {
		warnf("brew not found after install — skipping packages")
		return
	}
	logf("brew: %s", brew)
	mustRun(brew, "install", "zsh", "coreutils", "unzip", "openssl@3")
}

// brewPath returns the arch-native Homebrew binary if present.
// Apple Silicon uses /opt/homebrew; Intel Macs use /usr/local.
func brewPath() string {
	var candidates []string
	if runtime.GOARCH == "arm64" {
		candidates = []string{"/opt/homebrew/bin/brew", "/usr/local/bin/brew"}
	} else {
		candidates = []string{"/usr/local/bin/brew", "/opt/homebrew/bin/brew"}
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	if p, err := exec.LookPath("brew"); err == nil {
		return p
	}
	return ""
}

func installLinux() {
	logf("installing Linux packages...")
	mustRun("sudo", "apt-get", "update")
	mustRun("sudo", "apt-get", "install", "-y",
		"zsh", "build-essential", "unzip", "xdg-utils")
	// best-effort: not all Linux envs have xdg-settings
	run("xdg-settings", "set", "default-web-browser", "file-protocol-handler.desktop")
}

// ── YAML config parsing ────────────────────────────────────────────────────

type config struct {
	defaults linkOpts
	steps    []step
}

type step struct {
	kind  string // "create" | "link" | "shell"
	dirs  []string
	links []linkEntry
	cmds  []shellCmd
}

type linkEntry struct {
	target string
	spec   linkSpec
}

type linkSpec struct {
	path   string
	relink *bool
	force  *bool
}

type linkOpts struct {
	Relink bool `yaml:"relink"`
	Force  bool `yaml:"force"`
}

type shellCmd struct {
	Command     string `yaml:"command"`
	Description string `yaml:"description"`
}

func parseConfig(path string) (*config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	if doc.Kind != yaml.DocumentNode || len(doc.Content) == 0 {
		return nil, fmt.Errorf("empty config")
	}
	seq := doc.Content[0]
	if seq.Kind != yaml.SequenceNode {
		return nil, fmt.Errorf("config must be a YAML sequence")
	}

	cfg := &config{}
	for _, item := range seq.Content {
		if item.Kind != yaml.MappingNode || len(item.Content) < 2 {
			continue
		}
		key := item.Content[0].Value
		val := item.Content[1]

		switch key {
		case "defaults":
			var d struct {
				Link linkOpts `yaml:"link"`
			}
			if err := val.Decode(&d); err != nil {
				return nil, fmt.Errorf("defaults: %w", err)
			}
			cfg.defaults = d.Link

		case "create":
			var dirs []string
			if err := val.Decode(&dirs); err != nil {
				return nil, fmt.Errorf("create: %w", err)
			}
			cfg.steps = append(cfg.steps, step{kind: "create", dirs: dirs})

		case "link":
			entries, err := parseLinkEntries(val)
			if err != nil {
				return nil, fmt.Errorf("link: %w", err)
			}
			cfg.steps = append(cfg.steps, step{kind: "link", links: entries})

		case "shell":
			var cmds []shellCmd
			if err := val.Decode(&cmds); err != nil {
				return nil, fmt.Errorf("shell: %w", err)
			}
			cfg.steps = append(cfg.steps, step{kind: "shell", cmds: cmds})

		default:
			warnf("unknown directive %q, skipping", key)
		}
	}
	return cfg, nil
}

func parseLinkEntries(node *yaml.Node) ([]linkEntry, error) {
	if node.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("expected mapping")
	}
	entries := make([]linkEntry, 0, len(node.Content)/2)
	for i := 0; i+1 < len(node.Content); i += 2 {
		spec, err := parseLinkSpec(node.Content[i+1])
		if err != nil {
			return nil, fmt.Errorf("%s: %w", node.Content[i].Value, err)
		}
		entries = append(entries, linkEntry{target: node.Content[i].Value, spec: spec})
	}
	return entries, nil
}

func parseLinkSpec(node *yaml.Node) (linkSpec, error) {
	switch node.Kind {
	case yaml.ScalarNode:
		return linkSpec{path: node.Value}, nil
	case yaml.MappingNode:
		var obj struct {
			Path   string `yaml:"path"`
			Relink *bool  `yaml:"relink"`
			Force  *bool  `yaml:"force"`
		}
		if err := node.Decode(&obj); err != nil {
			return linkSpec{}, err
		}
		return linkSpec{path: obj.Path, relink: obj.Relink, force: obj.Force}, nil
	default:
		return linkSpec{}, fmt.Errorf("expected string or mapping")
	}
}

// ── execution ──────────────────────────────────────────────────────────────

func execute(cfg *config, baseDir string) {
	for _, s := range cfg.steps {
		switch s.kind {
		case "create":
			for _, dir := range s.dirs {
				dir = expandHome(dir)
				logf("create  %s", dir)
				if err := os.MkdirAll(dir, 0o755); err != nil {
					warnf("  %v", err)
				}
			}

		case "link":
			for _, e := range s.links {
				target := expandHome(e.target)
				src := filepath.Join(baseDir, e.spec.path)
				relink := boolOr(e.spec.relink, cfg.defaults.Relink)
				force := boolOr(e.spec.force, cfg.defaults.Force)
				logf("link    %s -> %s", target, src)
				if err := makeLink(target, src, relink, force); err != nil {
					warnf("  %v", err)
				}
			}

		case "shell":
			for _, cmd := range s.cmds {
				label := cmd.Description
				if label == "" {
					label = cmd.Command
				}
				logf("shell   %s", label)
				c := exec.Command("bash", "-c", cmd.Command)
				c.Dir = baseDir
				c.Stdout = os.Stdout
				c.Stderr = os.Stderr
				if err := c.Run(); err != nil {
					warnf("  %v", err)
				}
			}
		}
	}
}

func makeLink(target, src string, relink, force bool) error {
	info, err := os.Lstat(target)
	if err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			existing, _ := os.Readlink(target)
			// Normalize relative symlinks to absolute before comparing.
			if !filepath.IsAbs(existing) {
				existing = filepath.Join(filepath.Dir(target), existing)
			}
			if existing == src {
				return nil // already correct
			}
			if !relink {
				return fmt.Errorf("symlink exists -> %s (set relink: true to replace)", existing)
			}
			os.Remove(target)
		} else {
			if !force {
				return fmt.Errorf("target exists and is not a symlink (set force: true to overwrite)")
			}
			if err := os.RemoveAll(target); err != nil {
				return err
			}
		}
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	// Use a relative symlink so the same link works across environments
	// with different home directory paths (e.g. /Users/w5000 vs /home/appuser).
	relSrc, err := filepath.Rel(filepath.Dir(target), src)
	if err != nil {
		relSrc = src
	}
	return os.Symlink(relSrc, target)
}

// ── post-install setup ─────────────────────────────────────────────────────

// writeStarshipLocalEnv writes machine-local env overrides to .zshenv.local.
// This file is sourced at the end of .zshenv so it takes precedence.
func writeStarshipLocalEnv(dotfilesDir string) {
	p := filepath.Join(dotfilesDir, "files/.config/zsh/.zshenv.local")

	lines := []string{
		`export STARSHIP_CONFIG="$XDG_CONFIG_HOME/starship/host_starship.toml"`,
	}

	// When ~/.config/aquaproj-aqua can't be symlinked (e.g. restricted Codespaces env),
	// .zshenv sets AQUA_GLOBAL_CONFIG/AQUA_POLICY_CONFIG to a non-existent path.
	// Override them here so aqua proxy works in every shell session.
	home, _ := os.UserHomeDir()
	if _, err := os.Stat(filepath.Join(home, ".config/aquaproj-aqua")); err != nil {
		aquaDir := filepath.Join(dotfilesDir, "files/.config/aquaproj-aqua")
		lines = append(lines,
			fmt.Sprintf("export AQUA_GLOBAL_CONFIG=%q", filepath.Join(aquaDir, "aqua.yaml")),
			fmt.Sprintf("export AQUA_POLICY_CONFIG=%q", filepath.Join(aquaDir, "policy.yaml")),
		)
	}

	content := strings.Join(lines, "\n") + "\n"
	logf("write  %s", p)
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		warnf("write starship env: %v", err)
	}
}

// installMise installs mise via the official installer script, then runs
// `mise install` to provision all tool versions declared in config.
func installMise() {
	home, _ := os.UserHomeDir()
	miseBin := filepath.Join(home, ".local/bin/mise")
	if _, err := os.Stat(miseBin); os.IsNotExist(err) {
		logf("installing mise...")
		c := exec.Command("bash", "-c", "curl https://mise.run | sh")
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			warnf("install mise: %v", err)
			return
		}
	}
	logf("mise install")
	c := exec.Command(miseBin, "install")
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		warnf("mise install: %v", err)
	}
}

// gitSkipWorktree marks the machine-specific git config so local credential
// helper edits are never accidentally staged.
func gitSkipWorktree(dotfilesDir string) {
	logf("git skip-worktree files/.config/git/config")
	c := exec.Command("git", "update-index", "--skip-worktree", "files/.config/git/config")
	c.Dir = dotfilesDir
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		warnf("git update-index: %v", err)
	}
}

// setupWSL copies .wezterm.lua to a Windows user directory when running in WSL.
// Skips silently on non-Linux or non-WSL systems.
func setupWSL(dotfilesDir string) {
	if runtime.GOOS != "linux" {
		return
	}
	windowsUsers := "/mnt/c/Users"
	if _, err := os.Stat(windowsUsers); err != nil {
		return
	}

	src := filepath.Join(dotfilesDir, "files/.config/wezterm/.wezterm.lua")
	if _, err := os.Stat(src); err != nil {
		warnf("wezterm config not found: %s", src)
		return
	}

	entries, err := os.ReadDir(windowsUsers)
	if err != nil {
		warnf("read Windows users: %v", err)
		return
	}
	var users []string
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() && name != "Public" && name != "Default" && !strings.HasPrefix(name, ".") {
			users = append(users, name)
		}
	}
	if len(users) == 0 {
		warnf("no Windows users found in %s", windowsUsers)
		return
	}

	if !isTTY() {
		warnf("WSL detected but not interactive — skipping wezterm setup")
		return
	}

	fmt.Println("WSL detected. Copy .wezterm.lua to a Windows user directory.")
	for i, u := range users {
		fmt.Printf("  %d: %s\n", i+1, u)
	}
	fmt.Print("Enter number (0 to skip): ")

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	var choice int
	fmt.Sscan(scanner.Text(), &choice)
	if choice <= 0 || choice > len(users) {
		logf("skipping wezterm setup")
		return
	}

	dst := filepath.Join(windowsUsers, users[choice-1], ".wezterm.lua")
	data, err := os.ReadFile(src)
	if err != nil {
		warnf("read wezterm config: %v", err)
		return
	}
	if err := os.WriteFile(dst, data, 0o644); err != nil {
		warnf("write wezterm config to %s: %v", dst, err)
		return
	}
	logf("copied .wezterm.lua → %s", dst)
}

// linkConfigSubdirs is a fallback for environments where ~/.config can't be
// replaced wholesale (e.g. Codespaces where ~/.config/gh is a Docker bind mount).
// If ~/.config is already correctly symlinked, this is a no-op.
// Otherwise it links each entry in files/.config/ individually, skipping
// entries that already exist.
func linkConfigSubdirs(dotfilesDir string) {
	home, _ := os.UserHomeDir()
	configTarget := filepath.Join(home, ".config")
	configSrc := filepath.Join(dotfilesDir, "files/.config")

	// Already correctly linked — nothing to do.
	if link, err := os.Readlink(configTarget); err == nil && link == configSrc {
		return
	}

	// Not a symlink (real dir) — link individual entries.
	if info, err := os.Lstat(configTarget); err != nil || info.Mode()&os.ModeSymlink != 0 {
		return
	}

	entries, err := os.ReadDir(configSrc)
	if err != nil {
		warnf("read %s: %v", configSrc, err)
		return
	}
	for _, e := range entries {
		dst := filepath.Join(configTarget, e.Name())
		if _, err := os.Lstat(dst); err == nil {
			continue // already exists, skip
		}
		src := filepath.Join(configSrc, e.Name())
		logf("link    %s -> %s", dst, src)
		if err := os.Symlink(src, dst); err != nil {
			warnf("  %v", err)
		}
	}
}

// ── helpers ────────────────────────────────────────────────────────────────

func mustRun(name string, args ...string) {
	if err := run(name, args...); err != nil {
		warnf("%s: %v", name, err)
	}
}

func run(name string, args ...string) error {
	c := exec.Command(name, args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

func runInteractive(name string, args ...string) {
	c := exec.Command(name, args...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		warnf("%s: %v", name, err)
	}
}

func isTTY() bool {
	fi, err := os.Stdin.Stat()
	return err == nil && fi.Mode()&os.ModeCharDevice != 0
}

func boolOr(ptr *bool, fallback bool) bool {
	if ptr != nil {
		return *ptr
	}
	return fallback
}

func expandHome(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}

func logf(format string, args ...any)  { fmt.Printf(format+"\n", args...) }
func warnf(format string, args ...any) { fmt.Fprintf(os.Stderr, "warning: "+format+"\n", args...) }
func die(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "fatal: "+format+"\n", args...)
	os.Exit(1)
}
