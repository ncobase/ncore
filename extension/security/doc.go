// Package security provides security controls and sandboxing for extension loading
// and execution, including plugin signature verification and resource monitoring.
//
// This package protects against:
//   - Malicious plugin code execution
//   - Path traversal attacks
//   - Unsigned or tampered plugins
//   - Resource exhaustion (memory, CPU)
//   - Unauthorized file access
//   - Untrusted plugin sources
//
// # Sandbox Configuration
//
// Create a sandbox with security policies:
//
//	cfg := &config.SecurityConfig{
//	    EnableSandbox:       true,
//	    RequireSignature:    true,
//	    AllowUnsafe:         false,
//	    AllowedPaths:        []string{"/plugins", "/extensions"},
//	    BlockedExtensions:   []string{".exe", ".sh", ".bat"},
//	    TrustedSources:      []string{"github.com/myorg"},
//	}
//
//	sandbox := security.NewSandbox(cfg)
//
// # Plugin Validation
//
// Validate plugin before loading:
//
//	// Check plugin path is allowed
//	if err := sandbox.ValidatePluginPath("/plugins/my-plugin.so"); err != nil {
//	    log.Fatal("Invalid plugin path:", err)
//	}
//
//	// Verify plugin signature
//	if err := sandbox.ValidatePluginSignature("/plugins/my-plugin.so"); err != nil {
//	    log.Fatal("Signature verification failed:", err)
//	}
//
//	// Check plugin source is trusted
//	if err := sandbox.ValidatePluginSource("github.com/myorg/plugin"); err != nil {
//	    log.Fatal("Untrusted plugin source:", err)
//	}
//
// # Signature Verification
//
// The package uses SHA256 hash-based signature verification:
//
//	// Plugin signature file (.sig) should contain SHA256 hash
//	// Example: my-plugin.so.sig
//	//   contains: "abc123...def456"
//
//	// Generate signature for plugin
//	hash := sha256.Sum256(pluginBytes)
//	signature := hex.EncodeToString(hash[:])
//	os.WriteFile("plugin.so.sig", []byte(signature), 0644)
//
// Signatures are verified automatically during ValidatePluginSignature.
//
// # Resource Monitoring
//
// Monitor and limit plugin resource usage:
//
//	monitor := security.NewResourceMonitor(&config.PerformanceConfig{
//	    MaxMemoryMB:   512,
//	    MaxCPUPercent: 50,
//	})
//
//	// Check if plugin can be loaded
//	if err := monitor.CheckResourceLimits("my-plugin"); err != nil {
//	    log.Fatal("Insufficient resources:", err)
//	}
//
//	// Record plugin metrics
//	monitor.RecordPluginMetrics("my-plugin", &security.PluginMetrics{
//	    MemoryUsageMB:   45.2,
//	    CPUUsagePercent: 12.5,
//	    LoadTime:        150 * time.Millisecond,
//	    InitTime:        50 * time.Millisecond,
//	})
//
//	// Get plugin metrics
//	metrics := monitor.GetPluginMetrics("my-plugin")
//
// # Path Validation
//
// Prevent directory traversal attacks:
//
//	// Only allows access to whitelisted directories
//	cfg.AllowedPaths = []string{"/var/plugins", "/opt/extensions"}
//
//	// Blocks dangerous file extensions
//	cfg.BlockedExtensions = []string{".exe", ".sh", ".bat", ".cmd"}
//
// # Development vs Production
//
// Use AllowUnsafe for development only:
//
//	if os.Getenv("ENV") == "development" {
//	    cfg.AllowUnsafe = true      // Skip signature checks
//	    cfg.RequireSignature = false
//	} else {
//	    cfg.AllowUnsafe = false     // Enforce all security
//	    cfg.RequireSignature = true
//	}
//
// # Best Practices
//
//   - Always enable sandbox in production
//   - Require signatures for production plugins
//   - Use HTTPS for plugin distribution
//   - Limit plugin file access to specific directories
//   - Monitor resource usage continuously
//   - Implement plugin update verification
//   - Log all security validation failures
//   - Use trusted plugin sources only
//   - Set conservative resource limits
//   - Clean up metrics when plugins are removed
package security
