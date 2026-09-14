use serde::Deserialize;
use std::{fs, path::Path};

#[derive(Debug, Deserialize)]
pub struct SetIr {
    pub name: String,
}

#[derive(Debug, Deserialize)]
pub struct ModelIr {
    #[serde(rename = "schemaVersion")]
    pub schema_version: String,
    pub name: String,
    pub sets: Vec<SetIr>,
}

pub fn load_ir(path: impl AsRef<Path>) -> Result<ModelIr, Box<dyn std::error::Error>> {
    let data = fs::read_to_string(path)?;
    Ok(serde_json::from_str(&data)?)
}
