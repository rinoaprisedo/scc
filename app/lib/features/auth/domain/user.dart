class User {
  final String uuid;
  final String name;
  final String email;
  final String? avatar;
  final String status;
  final List<String> roles;
  final bool isSuperadmin;

  const User({
    required this.uuid,
    required this.name,
    required this.email,
    this.avatar,
    required this.status,
    required this.roles,
    required this.isSuperadmin,
  });

  factory User.fromJson(Map<String, dynamic> json) {
    return User(
      uuid: json['uuid'] as String? ?? '',
      name: json['name'] as String? ?? '',
      email: json['email'] as String? ?? '',
      avatar: json['avatar'] as String?,
      status: json['status'] as String? ?? '',
      roles: (json['roles'] as List?)?.map((e) => e.toString()).toList() ?? const [],
      isSuperadmin: json['is_superadmin'] as bool? ?? false,
    );
  }
}
