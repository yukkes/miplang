package io.github.yukkes.miplang.runtime;

import com.google.gson.Gson;
import com.google.gson.annotations.SerializedName;
import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.List;

public record ModelIr(
    @SerializedName("schemaVersion") String schemaVersion,
    String name,
    List<SetIr> sets) {

  private static final Gson GSON = new Gson();

  public static ModelIr load(Path path) throws IOException {
    return GSON.fromJson(Files.readString(path), ModelIr.class);
  }

  public record SetIr(String name) {}
}
