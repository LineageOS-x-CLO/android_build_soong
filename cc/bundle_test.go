// Copyright 2026 Google Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cc

import (
	"android/soong/android"
	"strings"
	"testing"
)

func TestBundleLibrary(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		ctx := testCc(t, `
		cc_library {
			name: "libfoo",
			srcs: ["foo.c"],
			system_shared_libs: [],
			split_all_variants: true,
			target_bundle: "libbundle",
		}

		cc_library_bundle {
			name: "libbundle",
			bundled_libs: ["libfoo"],
			system_shared_libs: [],
			split_all_variants: true,
		}

		cc_library {
			name: "libbar",
			srcs: ["bar.c"],
			shared_libs: ["libfoo"],
			system_shared_libs: [],
			split_all_variants: true,
		}`)

		libbarShared := ctx.ModuleForTests(t, "libbar", "android_arm64_armv8-a_shared").Rule("ld")
		libFlags := libbarShared.Args["libFlags"]
		if !strings.Contains(libFlags, "libbundle.so") {
			t.Errorf("expected libbundle.so in libbar shared libs, but not found in %q", libFlags)
		}
		if strings.Contains(libFlags, "libfoo.so") {
			t.Errorf("expected libfoo.so to be redirected to libbundle, but found in %q", libFlags)
		}

		libfoo := ctx.ModuleForTests(t, "libfoo", "android_arm64_armv8-a_shared").Module().(*Module)
		if !libfoo.Properties.PreventInstall {
			t.Errorf("expected PreventInstall to be true for libfoo")
		}
	})

	t.Run("bundleIntoSingle", func(t *testing.T) {
		ctx := testCc(t, `
		cc_library {
			name: "libuser",
			shared_libs: ["libfoo", "libbar"],
			system_shared_libs: [],
			split_all_variants: true,
		}

		cc_library_bundle {
			name: "libbundle",
			bundled_libs: ["libfoo", "libbar"],
			system_shared_libs: [],
			split_all_variants: true,
		}

		cc_library {
			name: "libfoo",
			system_shared_libs: [],
			split_all_variants: true,
			target_bundle: "libbundle",
		}

		cc_library {
			name: "libbar",
			system_shared_libs: [],
			split_all_variants: true,
			target_bundle: "libbundle",
		}`)

		libuserShared := ctx.ModuleForTests(t, "libuser", "android_arm64_armv8-a_shared").Rule("ld")
		libFlags := libuserShared.Args["libFlags"]
		if !strings.Contains(libFlags, "libbundle.so") {
			t.Errorf("expected libbundle.so in libuser shared libs, but not found in %q", libFlags)
		}
		if strings.Contains(libFlags, "libfoo.so") {
			t.Errorf("expected libfoo.so to be redirected to libbundle, but found in %q", libFlags)
		}
		if strings.Contains(libFlags, "libbar.so") {
			t.Errorf("expected libbar.so to be redirected to libbundle, but found in %q", libFlags)
		}

		libfoo := ctx.ModuleForTests(t, "libfoo", "android_arm64_armv8-a_shared").Module().(*Module)
		if !libfoo.Properties.PreventInstall {
			t.Errorf("expected PreventInstall to be true for libfoo")
		}
		libbar := ctx.ModuleForTests(t, "libbar", "android_arm64_armv8-a_shared").Module().(*Module)
		if !libbar.Properties.PreventInstall {
			t.Errorf("expected PreventInstall to be true for libbar")
		}
	})

	t.Run("multipleBundles", func(t *testing.T) {
		ctx := testCc(t, `
		cc_library {
			name: "libuser",
			shared_libs: ["libfoo", "libbar"],
			system_shared_libs: [],
			split_all_variants: true,
		}

		cc_library_bundle {
			name: "libbundleA",
			bundled_libs: ["libfoo"],
			shared_libs: ["libbar"],
			system_shared_libs: [],
			split_all_variants: true,
		}

		cc_library_bundle {
			name: "libbundleB",
			bundled_libs: ["libbar"],
			system_shared_libs: [],
			split_all_variants: true,
		}

		cc_library {
			name: "libfoo",
			shared_libs: ["libbar"],
			system_shared_libs: [],
			split_all_variants: true,
			target_bundle: "libbundleA",
		}

		cc_library {
			name: "libbar",
			system_shared_libs: [],
			split_all_variants: true,
			target_bundle: "libbundleB",
		}`)

		{
			libuserShared := ctx.ModuleForTests(t, "libuser", "android_arm64_armv8-a_shared").Rule("ld")
			libFlags := libuserShared.Args["libFlags"]
			if !strings.Contains(libFlags, "libbundleA.so") {
				t.Errorf("expected libbundleA.so in libuser shared libs, but not found in %q", libFlags)
			}
			if !strings.Contains(libFlags, "libbundleB.so") {
				t.Errorf("expected libbundleB.so in libuser shared libs, but not found in %q", libFlags)
			}
			if strings.Contains(libFlags, "libfoo.so") {
				t.Errorf("expected libfoo.so to be redirected to libbundleA, but found in %q", libFlags)
			}
			if strings.Contains(libFlags, "libbar.so") {
				t.Errorf("expected libbar.so to be redirected to libbundleB, but found in %q", libFlags)
			}
		}

		{
			libbundleAShared := ctx.ModuleForTests(t, "libbundleA", "android_arm64_armv8-a_shared").Rule("ld")
			libFlags := libbundleAShared.Args["libFlags"]
			if !strings.Contains(libFlags, "libbundleB.so") {
				t.Errorf("expected libbundleB.so in libbundleA shared libs, but not found in %q", libFlags)
			}
			if strings.Contains(libFlags, "libbar.so") {
				t.Errorf("expected libbar.so to be redirected to libbundleB, but found in %q", libFlags)
			}
		}

		libfoo := ctx.ModuleForTests(t, "libfoo", "android_arm64_armv8-a_shared").Module().(*Module)
		if !libfoo.Properties.PreventInstall {
			t.Errorf("expected PreventInstall to be true for libfoo")
		}
		libbar := ctx.ModuleForTests(t, "libbar", "android_arm64_armv8-a_shared").Module().(*Module)
		if !libbar.Properties.PreventInstall {
			t.Errorf("expected PreventInstall to be true for libbar")
		}
	})

	t.Run("no redirection for vendor", func(t *testing.T) {
		ctx := testCc(t, `
		cc_library {
			name: "libfoo",
			srcs: ["foo.c"],
			system_shared_libs: [],
			vendor_available: true,
			split_all_variants: true,
			target_bundle: "libbundle",
		}

		cc_library_bundle {
			name: "libbundle",
			bundled_libs: ["libfoo"],
			system_shared_libs: [],
			split_all_variants: true,
		}

		cc_library {
			name: "libbar",
			srcs: ["bar.c"],
			shared_libs: ["libfoo"],
			system_shared_libs: [],
			vendor_available: true,
			split_all_variants: true,
		}`)

		libbarVendorShared := ctx.ModuleForTests(t, "libbar", "android_vendor_arm64_armv8-a_shared").Rule("ld")
		libFlags := libbarVendorShared.Args["libFlags"]
		if strings.Contains(libFlags, "libbundle.so") {
			t.Errorf("expected libbundle.so NOT in libbar vendor shared libs, but found in %q", libFlags)
		}
		if !strings.Contains(libFlags, "libfoo.so") {
			t.Errorf("expected libfoo.so in libbar vendor shared libs, but not found in %q", libFlags)
		}

		libfooVendor := ctx.ModuleForTests(t, "libfoo", "android_vendor_arm64_armv8-a_shared").Module().(*Module)
		if libfooVendor.Properties.PreventInstall {
			t.Errorf("expected PreventInstall to be false for libfoo vendor variant")
		}
	})

	t.Run("duplicate bundle error", func(t *testing.T) {
		testCcError(t, `bundled in a different module`, `
		cc_library {
			name: "libfoo",
			srcs: ["foo.c"],
			system_shared_libs: [],
			split_all_variants: true,
			target_bundle: "libbundle1",
		}

		cc_library_bundle {
			name: "libbundle1",
			bundled_libs: ["libfoo"],
			system_shared_libs: [],
		}

		cc_library_bundle {
			name: "libbundle2",
			bundled_libs: ["libfoo"],
			system_shared_libs: [],
		}`)
	})

	t.Run("target_bundle missing error", func(t *testing.T) {
		testCcError(t, "`target_bundle: \"libbundle\"` is missing", `
		cc_library {
			name: "libfoo",
			srcs: ["foo.c"],
			system_shared_libs: [],
			split_all_variants: true,
		}

		cc_library_bundle {
			name: "libbundle",
			bundled_libs: ["libfoo"],
			system_shared_libs: [],
		}`)
	})

	t.Run("dependency among bundled libs", func(t *testing.T) {
		ctx := testCc(t, `
		cc_library {
			name: "libfoo",
			srcs: ["foo.c"],
			shared_libs: ["libbar"],
			system_shared_libs: [],
			split_all_variants: true,
			target_bundle: "libbundle",
		}

		cc_library {
			name: "libbar",
			srcs: ["bar.c"],
			system_shared_libs: [],
			split_all_variants: true,
			target_bundle: "libbundle",
		}

		cc_library_bundle {
			name: "libbundle",
			bundled_libs: ["libfoo", "libbar"],
			system_shared_libs: [],
			split_all_variants: true,
		}

		cc_library {
			name: "libuser",
			srcs: ["user.c"],
			shared_libs: ["libfoo"],
			system_shared_libs: [],
			split_all_variants: true,
		}`)

		libuserShared := ctx.ModuleForTests(t, "libuser", "android_arm64_armv8-a_shared").Rule("ld")
		libFlags := libuserShared.Args["libFlags"]
		if !strings.Contains(libFlags, "libbundle.so") {
			t.Errorf("expected libbundle.so in libuser shared libs, but not found in %q", libFlags)
		}
	})

	t.Run("unpicked library with target_bundle", func(t *testing.T) {
		// Test that a library setting target_bundle but not picked by the bundle's bundled_libs
		// does not cause an error. This happens with multi-version AIDL libs.
		testCc(t, `
		cc_library {
			name: "libfoo",
			srcs: ["foo.c"],
			system_shared_libs: [],
			split_all_variants: true,
			target_bundle: "libbundle",
		}

		cc_library {
			name: "libbar",
			srcs: ["bar.c"],
			system_shared_libs: [],
			split_all_variants: true,
			target_bundle: "libbundle",
		}

		cc_library_bundle {
			name: "libbundle",
			bundled_libs: ["libfoo"], // libbar is not picked
			system_shared_libs: [],
			split_all_variants: true,
		}

		cc_library {
			name: "libuser",
			srcs: ["user.c"],
			shared_libs: ["libfoo", "libbar"],
			system_shared_libs: [],
			split_all_variants: true,
		}`)
	})
}

func TestBundleLibrary_IncludeDir(t *testing.T) {
	t.Parallel()
	ctx := android.GroupFixturePreparers(
		prepareForCcTest,
		android.MockFS{
			"user/Android.bp": []byte(`
				cc_binary {
					name: "user",
					srcs: ["user.c"],
					shared_libs: ["libfoo"],
					system_shared_libs: [],
				}
			`),
			"bundle/Android.bp": []byte(`
				cc_library_bundle {
					name: "libbundle",
					bundled_libs: ["libfoo", "libbar"],
					system_shared_libs: [],
					split_all_variants: true,
				}
			`),
			"foo/Android.bp": []byte(`
				cc_library {
					name: "libfoo",
					target_bundle: "libbundle",
					export_include_dirs: ["inc"],
					system_shared_libs: [],
					split_all_variants: true,
				}
			`),
			"bar/Android.bp": []byte(`
				cc_library {
					name: "libbar",
					target_bundle: "libbundle",
					export_include_dirs: ["inc"],
					system_shared_libs: [],
					split_all_variants: true,
				}
			`),
		}.AddToFixture(),
	).RunTest(t).TestContext

	user := ctx.ModuleForTests(t, "user", "android_arm64_armv8-a")
	userCcRule := user.Rule("cc")
	userCcFlags := userCcRule.Args["cFlags"]
	android.AssertStringDoesContain(t, "should include foo/inc", userCcFlags, "-Ifoo/inc")
	android.AssertStringDoesNotContain(t, "should not include bar/inc", userCcFlags, "-Ibar/inc")
}

func TestBundleLibrary_ReExportFromBundledLibs(t *testing.T) {
	t.Parallel()
	ctx := android.GroupFixturePreparers(
		prepareForCcTest,
		android.MockFS{
			"user/Android.bp": []byte(`
				cc_binary {
					name: "user",
					srcs: ["user.c"],
					shared_libs: ["libfoo"],
					system_shared_libs: [],
				}
			`),
			"bundle/Android.bp": []byte(`
				cc_library_bundle {
					name: "libbundle",
					bundled_libs: ["libbar"],
					system_shared_libs: [],
					split_all_variants: true,
				}
			`),
			"foo/Android.bp": []byte(`
				cc_library {
					name: "libfoo",
					srcs: ["foo.c"],
					export_include_dirs: ["inc"],
					shared_libs: ["libbar"],
					export_shared_lib_headers: ["libbar"],
					system_shared_libs: [],
					split_all_variants: true,
				}
			`),
			"bar/Android.bp": []byte(`
				cc_library {
					name: "libbar",
					srcs: ["bar.c"],
					target_bundle: "libbundle",
					export_include_dirs: ["inc"],
					system_shared_libs: [],
					split_all_variants: true,
				}
			`),
		}.AddToFixture(),
	).RunTest(t).TestContext

	// user should get libbar via libfoo:
	//
	// user --> libfoo ----------> libbar
	//                 shared_libs
	//                  re-export
	//
	// libfoo should re-export the libbar's include_dirs even when it's replaced
	// with a bundle.
	user := ctx.ModuleForTests(t, "user", "android_arm64_armv8-a")
	userCcRule := user.Rule("cc")
	userCcFlags := userCcRule.Args["cFlags"]
	android.AssertStringDoesContain(t, "should include bar/inc", userCcFlags, "-Ibar/inc")
}
