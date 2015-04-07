# The files DEPS and .DEPS.git need to be manually kept in sync. Depending on
# whether buildtools is used from a svn or git project one or the other is used.

recursion = 1
use_relative_paths = True

vars = {
  "git_url": "https://chromium.googlesource.com",

  "clang_format_rev": "81edd558fea5dd7855d67a1dc61db34ae8c1fd63", # r223685
  "libcxx_revision": "48198f9110397fff47fe7c37cbfa296be7d44d3d",
  "libcxxabi_revision": "4ad1009ab3a59fa7a6896d74d5e4de5885697f95",
}

deps = {
  "clang_format/script":
      Var("git_url") + "/chromium/llvm-project/cfe/tools/clang-format.git@" +
      Var("clang_format_rev"),
  "third_party/libc++/trunk":
      Var("git_url") + "/chromium/llvm-project/libcxx.git" + "@" +
      Var("libcxx_revision"),
  "third_party/libc++abi/trunk":
      Var("git_url") + "/chromium/llvm-project/libcxxabi.git" + "@" +
      Var("libcxxabi_revision"),
}
