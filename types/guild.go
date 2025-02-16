package types

type GuildFetch struct {
	Guild string `json:"guild"`
}

type GetMutualGuilds struct {
	Guilds []MutualGuild `json:"guilds"`
}

type MutualGuild struct {
	Name            string `json:"name"`
	ID              string `json:"id"`
	Icon            string `json:"icon_url"`
	MemberCount     string `json:"member_count"`
	PermissionLevel int    `json:"permission_level"`
}
