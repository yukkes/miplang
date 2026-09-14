package io.github.yukkes.miplang.runtime;

import static org.junit.jupiter.api.Assertions.assertEquals;

import java.nio.file.Path;
import org.junit.jupiter.api.Test;

class ModelIrTest {
  @Test
  void loadsSharedIr() throws Exception {
    var model = ModelIr.load(Path.of("../../testdata/transport.ir.json"));
    assertEquals("miplang.ir/v1alpha2", model.schemaVersion());
    assertEquals("anonymous", model.name());
    assertEquals("TRIPS", model.sets().getFirst().name());
  }
}
