use miplang_runtime::load_ir;

#[test]
fn loads_shared_ir() {
    let model = load_ir("../../testdata/transport.ir.json").unwrap();
    assert_eq!(model.schema_version, "miplang.ir/v1alpha2");
    assert_eq!(model.name, "anonymous");
    assert_eq!(model.sets[0].name, "TRIPS");
}
