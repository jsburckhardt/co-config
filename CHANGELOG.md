# Changelog

## [0.1.2](https://github.com/jsburckhardt/co-config/compare/v0.1.1...v0.1.2) (2026-03-05)


### Bug Fixes

* **ci:** chain GoReleaser into release-please workflow ([1cb7016](https://github.com/jsburckhardt/co-config/commit/1cb7016d9f0ee4f39aceba6ca624225a95914cb6))
* **ci:** chain GoReleaser into release-please workflow (WI-0008) ([afdfb15](https://github.com/jsburckhardt/co-config/commit/afdfb1536b83777b6a92a6858b998e64f938371f))

## [0.1.1](https://github.com/jsburckhardt/co-config/compare/v0.1.0...v0.1.1) (2026-03-05)


### Features

* add ccc CLI entrypoint ([d2209b8](https://github.com/jsburckhardt/co-config/commit/d2209b869694f175da5c2e7b2ede997a4d518ee1))
* add CI/CD pipeline with GoReleaser, cosign signing, and release-please (WI-0008) ([07f2565](https://github.com/jsburckhardt/co-config/commit/07f2565582e39b71cb523cdcfaee2a673298df5b))
* add Copilot CLI integration tests ([8221b6c](https://github.com/jsburckhardt/co-config/commit/8221b6c030c304d2d786025a266d8096da2adeed))
* add copilot detection, TUI, justfile, and doc fixes [WI-0002] ([e65e3f2](https://github.com/jsburckhardt/co-config/commit/e65e3f2614acafc26160a14808eade1a89235d96))
* add curl-based install script and security policy ([792c4e6](https://github.com/jsburckhardt/co-config/commit/792c4e6e6b8eb64763dffc9cab846cdd67b16229))
* bootstrap ccc project with Go + Charm TUI stack ([30f0492](https://github.com/jsburckhardt/co-config/commit/30f0492c92d4c4a51ebb46f6dee4e84ce823e1df))
* bootstrap ccc project with Go + Charm TUI stack ([953b090](https://github.com/jsburckhardt/co-config/commit/953b0902ae908bdf6e8a50a02fbc09a94c5869e4))
* implement configuration management ([2f564d9](https://github.com/jsburckhardt/co-config/commit/2f564d9dde295c1c6ac2691c285d31882520aecc))
* implement logging infrastructure ([a14e411](https://github.com/jsburckhardt/co-config/commit/a14e411cf587abb9b5dae510967befcefe5c2aec))
* implement sensitive data handling ([3ab25eb](https://github.com/jsburckhardt/co-config/commit/3ab25eb0db31bd1b4cb8245c1c4aa09508f3f72b))
* **tui:** add environment variables view with multi-view navigation ([ed4e17b](https://github.com/jsburckhardt/co-config/commit/ed4e17b78fa03fbb9903b0fe45b1190d1d7a0bbb))
* **tui:** add environment variables view with multi-view navigation ([aaf38ca](https://github.com/jsburckhardt/co-config/commit/aaf38ca40cd0800a25e70d57b3b432206bb7479f))
* **tui:** add filterable model picker for large enum fields ([b1a3070](https://github.com/jsburckhardt/co-config/commit/b1a307037365f13168d54a561e4847c30a8a9f45))
* **tui:** add filterable model picker for large enum fields ([23a2d73](https://github.com/jsburckhardt/co-config/commit/23a2d736c64ab535bf68b5518c2fedc8b87213b3))
* **tui:** improve header/help UI with Copilot icon and show default values ([f8767fa](https://github.com/jsburckhardt/co-config/commit/f8767fa7b0a84d7d4b0c232f1e4f65d08fcea5bb))
* **tui:** improve header/help UI with Copilot icon and show default values ([c96924d](https://github.com/jsburckhardt/co-config/commit/c96924d022824bd19ebbe824971f8201016d73fd))
* two-panel TUI redesign (WI-0003) ([07ff458](https://github.com/jsburckhardt/co-config/commit/07ff4584e67549012bd23739323f55cd16d2588f))


### Bug Fixes

* Add error handling to DefaultPath with temp dir fallback ([e3d0d4d](https://github.com/jsburckhardt/co-config/commit/e3d0d4d03c1cbe75c6e9f93029c7616e6048eb49))
* address all 10 Copilot PR review comments [WI-0002] ([47245b7](https://github.com/jsburckhardt/co-config/commit/47245b73bb5d3df805e8228a4c976c8f1eff41d1))
* **ci:** correct golangci-lint-action SHA pin ([f6f8615](https://github.com/jsburckhardt/co-config/commit/f6f86154e1a6a36c3c45d74fb1b0eaf67c14fb6d))
* **ci:** correct SHA pins for release-please, codeql, cosign, and syft actions ([48e9eb5](https://github.com/jsburckhardt/co-config/commit/48e9eb543b73fe1926869f5784ddeefeeff6a028))
* **ci:** correct SHA pins for release-please, codeql, cosign, and syft actions (WI-0008) ([84d026a](https://github.com/jsburckhardt/co-config/commit/84d026adabfd75ab3d5e356a05a7569ccd2210fd))
* correct capitalization of "Copilot" in PR assignment instructions ([0421729](https://github.com/jsburckhardt/co-config/commit/042172981501e4ddd85227ad68143f4695b61f4b))
* Handle Close() error in logging.Init [fix-close-error] ([b95ac23](https://github.com/jsburckhardt/co-config/commit/b95ac2331e7895a0e01d9c0274473356769e3abc))
* Replace fragile /root/forbidden test with chmod-based approach ([fd10cb0](https://github.com/jsburckhardt/co-config/commit/fd10cb05d1aa24151583cf2ac7d6b537677305e8))
* rewrite TUI rendering — fix double borders, panel height, layout ([f4f67fb](https://github.com/jsburckhardt/co-config/commit/f4f67fbb99ad2d019e4fd67136fe51ab7cadc342))
