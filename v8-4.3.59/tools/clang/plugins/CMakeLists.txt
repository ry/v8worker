add_llvm_loadable_module(libFindBadConstructs
                         ChromeClassTester.cpp
                         FindBadConstructsAction.cpp
                         FindBadConstructsConsumer.cpp
                        )

add_dependencies(libFindBadConstructs clang)

cr_install(TARGETS libFindBadConstructs LIBRARY DESTINATION lib)

cr_add_test(plugins_test
  ${CMAKE_CURRENT_SOURCE_DIR}/tests/test.sh
  ${CMAKE_BINARY_DIR}/bin/clang
  $<TARGET_FILE:libFindBadConstructs>
  )
