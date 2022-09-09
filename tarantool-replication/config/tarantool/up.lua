box.cfg{listen=3301}
box.schema.space.create('binlog',{id=512, if_not_exists = true})
box.space.binlog:create_index('primary', {
	type = 'HASH',
	if_not_exists = true,
	parts = {1, 'unsigned'}
});



box.schema.space.create('users',{id=513, if_not_exists = true, field_count = 7})
box.space.users:create_index('primary', {
	type = 'HASH',
	if_not_exists = true,
	parts = {1, 'unsigned'}
});
box.space.users:create_index('login', {
	unique = true,
	if_not_exists = true,
	parts = {2, 'string'}
});
box.space.users:create_index('email', {
	unique = false,
	if_not_exists = true,
	parts = {3, 'string'}
});
