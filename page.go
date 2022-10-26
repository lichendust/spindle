package main

type page_object struct {
	page_id   uint32
	page_path string
	content   []ast_data
	top_scope []*ast_declare
}