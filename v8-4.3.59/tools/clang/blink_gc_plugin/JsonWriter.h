// Copyright 2014 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

#ifndef TOOLS_BLINK_GC_PLUGIN_JSON_WRITER_H_
#define TOOLS_BLINK_GC_PLUGIN_JSON_WRITER_H_

#include "llvm/Support/raw_ostream.h"

// Helper to write information for the points-to graph.
class JsonWriter {
 public:
  static JsonWriter* from(llvm::raw_fd_ostream* os) {
    return os ? new JsonWriter(os) : 0;
  }
  ~JsonWriter() {
    os_.close();
  }
  void OpenList() {
    Separator();
    os_ << "[";
    state_.push(false);
  }
  void OpenList(const std::string key) {
    Write(key);
    os_ << ":";
    OpenList();
  }
  void CloseList() {
    os_ << "]";
    state_.pop();
  }
  void OpenObject() {
    Separator();
    os_ << "{";
    state_.push(false);
  }
  void CloseObject() {
    os_ << "}\n";
    state_.pop();
  }
  void Write(const size_t val) {
    Separator();
    os_ << val;
  }
  void Write(const std::string val) {
    Separator();
    os_ << "\"" << val << "\"";
  }
  void Write(const std::string key, const size_t val) {
    Separator();
    os_ << "\"" << key << "\":" << val;
  }
  void Write(const std::string key, const std::string val) {
    Separator();
    os_ << "\"" << key << "\":\"" << val << "\"";
  }
 private:
  JsonWriter(llvm::raw_fd_ostream* os) : os_(*os) {}
  void Separator() {
    if (state_.empty())
      return;
    if (state_.top()) {
      os_ << ",";
      return;
    }
    state_.top() = true;
  }
  llvm::raw_fd_ostream& os_;
  std::stack<bool> state_;
};

#endif // TOOLS_BLINK_GC_PLUGIN_JSON_WRITER_H_
