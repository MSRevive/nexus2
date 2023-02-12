package service_test

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/msrevive/nexus2/ent"
	entCharacter "github.com/msrevive/nexus2/ent/character"
	entPlayer "github.com/msrevive/nexus2/ent/player"
	"github.com/msrevive/nexus2/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCharactersGetAll_WithNoCharacters_ReturnsEmptySlice(t *testing.T) {
	refreshDb()
	
	ctx := context.Background()
	actual, err := service.New(ctx, testApp).CharactersGetAll()
	require.NoError(t, err, "unexpected error returned: %v", err)

	assert.NotNil(t, actual, "Characters slice should not be nil")
	assert.Empty(t, actual, "Characters slice should be empty")
}

func TestCharactersGetAll_WithSoftDeletedCharacters_ReturnsEmptySlice(t *testing.T) {
	refreshDb()

	ctx := context.Background()
	player := seedPlayer(t, ctx)
	now := time.Now()
	seedCharacterWithData(t, ctx, player, &ent.Character{Version: 1, DeletedAt: &now})

	actual, err := service.New(ctx, testApp).CharactersGetAll()
	require.NoError(t, err, "unexpected error returned: %v", err)

	assert.NotNil(t, actual, "Characters slice should not be nil")
	assert.Empty(t, actual, "Characters slice should be empty")
}

func TestCharactersGetAll_WithOneCharacters_ReturnsOneCharacter(t *testing.T) {
	refreshDb()

	ctx := context.Background()
	player := seedPlayer(t, ctx)
	character := seedCharacter(t, ctx, player)

	actual, err := service.New(ctx, testApp).CharactersGetAll()
	require.NoError(t, err, "unexpected error when calling CharactersGetAll")

	require.Len(t, actual, 1)
	assert.Equal(t, character.ID, actual[0].ID)
	assert.Equal(t, player.Steamid, actual[0].Steamid)
	assert.Equal(t, character.Slot, actual[0].Slot)
	assert.Equal(t, 1, character.Version)
	assert.Equal(t, character.Size, actual[0].Size)
	assert.Equal(t, character.Data, actual[0].Data)
}

func TestCharactersGetAll_WithMultipleCharacters_ReturnsAllCharacters(t *testing.T) {
	refreshDb()

	ctx := context.Background()
	player := seedPlayer(t, ctx)
	characters := seedCharacters(t, ctx, player, 3, 1)
	require.Len(t, characters, 3)

	actual, err := service.New(ctx, testApp).CharactersGetAll()
	require.NoError(t, err, "unexpected error when calling CharactersGetAll")

	assert.Len(t, actual, 3)
	for _, actualCharacter := range actual {
		character := charById(characters, actualCharacter.ID)
		require.NotNil(t, character)

		assert.Equal(t, character.ID, actualCharacter.ID)
		assert.Equal(t, player.Steamid, actualCharacter.Steamid)
		assert.Equal(t, character.Slot, actualCharacter.Slot)
		assert.Equal(t, 1, character.Version)
		assert.Equal(t, character.Size, actualCharacter.Size)
		assert.Equal(t, character.Data, actualCharacter.Data)
	}
}

func TestCharactersGetAll_WithMultipleCharactersAndVersions_ReturnsAllCharactersLatestVersions(t *testing.T) {
	refreshDb()

	ctx := context.Background()
	player := seedPlayer(t, ctx)
	characters := seedCharacters(t, ctx, player, 3, 3)
	require.Len(t, characters, 9)

	actual, err := service.New(ctx, testApp).CharactersGetAll()
	require.NoError(t, err, "unexpected error when calling CharactersGetAll")

	assert.Len(t, actual, 3)
	for _, actualCharacter := range actual {
		character := charById(characters, actualCharacter.ID)
		require.NotNil(t, character)

		assert.Equal(t, character.ID, actualCharacter.ID)
		assert.Equal(t, player.Steamid, actualCharacter.Steamid)
		assert.Equal(t, character.Slot, actualCharacter.Slot)
		assert.Equal(t, character.Size, actualCharacter.Size)
		assert.Equal(t, character.Data, actualCharacter.Data)
	}
}

func TestCharactersGetBySteamid_WithNoMatchingPlayer_ReturnsEmptySlice(t *testing.T) {
	ctx := context.Background()
	actual, err := service.New(ctx, testApp).CharactersGetBySteamid(uuid.NewString())
	require.NoError(t, err, "unexpected error returned: %v", err)

	assert.NotNil(t, actual, "Characters slice should not be nil")
	assert.Empty(t, actual, "Characters slice should be empty")
}

func TestCharactersGetBySteamid_WithSoftDeletedCharacters_ReturnsEmptySlice(t *testing.T) {
	refreshDb()

	ctx := context.Background()
	player := seedPlayer(t, ctx)
	now := time.Now()
	seedCharacterWithData(t, ctx, player, &ent.Character{Version: 1, DeletedAt: &now})

	actual, err := service.New(ctx, testApp).CharactersGetBySteamid(uuid.NewString())
	require.NoError(t, err, "unexpected error returned: %v", err)

	assert.NotNil(t, actual, "Characters slice should not be nil")
	assert.Empty(t, actual, "Characters slice should be empty")
}

func TestCharactersGetBySteamid_WithMatchingPlayer_AndNoCharacters_ReturnsEmptySlice(t *testing.T) {
	refreshDb()

	ctx := context.Background()
	player := seedPlayer(t, ctx)

	actual, err := service.New(ctx, testApp).CharactersGetBySteamid(player.Steamid)
	require.NoError(t, err, "unexpected error returned: %v", err)

	assert.NotNil(t, actual, "Characters slice should not be nil")
	assert.Empty(t, actual, "Characters slice should be empty")
}

func TestCharactersGetBySteamid_WithMatchingPlayer_ReturnsCharacters(t *testing.T) {
	refreshDb()

	ctx := context.Background()
	player := seedPlayer(t, ctx)
	character := seedCharacter(t, ctx, player)

	actual, err := service.New(ctx, testApp).CharactersGetBySteamid(player.Steamid)
	require.NoError(t, err, "unexpected error when calling CharactersGetBySteamid")

	require.Len(t, actual, 1)
	assert.Equal(t, character.ID, actual[0].ID)
	assert.Equal(t, player.Steamid, actual[0].Steamid)
	assert.Equal(t, character.Slot, actual[0].Slot)
	assert.Equal(t, 1, character.Version)
	assert.Equal(t, character.Size, actual[0].Size)
	assert.Equal(t, character.Data, actual[0].Data)
}

func TestCharactersGetBySteamid_WithMatchingPlayer_AndMultipleVersions_ReturnsLatestCharacterVersions(t *testing.T) {
	refreshDb()

	ctx := context.Background()
	player := seedPlayer(t, ctx)
	characters := seedCharacters(t, ctx, player, 3, 3)
	require.Len(t, characters, 9)

	actual, err := service.New(ctx, testApp).CharactersGetBySteamid(player.Steamid)
	require.NoError(t, err, "unexpected error when calling CharactersGetBySteamid")

	assert.Len(t, actual, 3)
	for _, actualCharacter := range actual {
		character := charById(characters, actualCharacter.ID)
		require.NotNil(t, character)

		assert.Equal(t, character.ID, actualCharacter.ID)
		assert.Equal(t, player.Steamid, actualCharacter.Steamid)
		assert.Equal(t, character.Slot, actualCharacter.Slot)
		assert.Equal(t, character.Size, actualCharacter.Size)
		assert.Equal(t, character.Data, actualCharacter.Data)
	}
}

func TestCharacterGetBySteamidSlot_WithNoMatchingPlayer_ReturnsNotFoundError(t *testing.T) {
	ctx := context.Background()
	actual, err := service.New(ctx, testApp).CharacterGetBySteamidSlot(uuid.NewString(), 0)
	assert.Nil(t, actual)
	assert.True(t, ent.IsNotFound(err))
}

func TestCharacterGetBySteamidSlot_WithMatchingPlayer_AndNoMatchingSlot_ReturnsNotFoundError(t *testing.T) {
	refreshDb()

	ctx := context.Background()
	player := seedPlayer(t, ctx)
	seedCharacter(t, ctx, player)

	actual, err := service.New(ctx, testApp).CharacterGetBySteamidSlot(player.Steamid, 3)
	assert.Nil(t, actual)
	assert.True(t, ent.IsNotFound(err))
}

func TestCharacterGetBySteamidSlot_WithSoftDeletedCharacter_ReturnsNotFoundError(t *testing.T) {
	refreshDb()

	ctx := context.Background()
	player := seedPlayer(t, ctx)
	now := time.Now()
	seedCharacterWithData(t, ctx, player, &ent.Character{Version: 1, DeletedAt: &now})

	actual, err := service.New(ctx, testApp).CharacterGetBySteamidSlot(player.Steamid, 0)
	assert.Nil(t, actual)
	assert.True(t, ent.IsNotFound(err))
}

func TestCharacterGetBySteamidSlot_WithMatchingPlayer_AndMatchingSlot_ReturnsCharacter(t *testing.T) {
	refreshDb()

	ctx := context.Background()
	player := seedPlayer(t, ctx)
	character := seedCharacter(t, ctx, player)

	actual, err := service.New(ctx, testApp).CharacterGetBySteamidSlot(player.Steamid, 0)
	require.NoError(t, err)

	assert.Equal(t, character.ID, actual.ID)
	assert.Equal(t, player.Steamid, actual.Steamid)
	assert.Equal(t, character.Slot, actual.Slot)
	assert.Equal(t, character.Size, actual.Size)
	assert.Equal(t, character.Data, actual.Data)
}

func TestCharacterGetBySteamidSlot_WithMatchingPlayer_AndMatchingSlot_WithMultipleVersions_ReturnsLatestCharacter(t *testing.T) {
	refreshDb()

	ctx := context.Background()
	player := seedPlayer(t, ctx)
	characters := seedCharacters(t, ctx, player, 1, 3)
	require.Len(t, characters, 3)

	actual, err := service.New(ctx, testApp).CharacterGetBySteamidSlot(player.Steamid, 0)
	require.NoError(t, err)

	character := charById(characters, actual.ID)
	require.NotNil(t, character)

	assert.Equal(t, character.ID, actual.ID)
	assert.Equal(t, player.Steamid, actual.Steamid)
	assert.Equal(t, 1, character.Version)
	assert.Equal(t, character.Slot, actual.Slot)
	assert.Equal(t, character.Size, actual.Size)
	assert.Equal(t, character.Data, actual.Data)
}

func TestCharacterGetByID_WithNoCharacters_ReturnsNotFoundError(t *testing.T) {
	refreshDb()

	ctx := context.Background()
	actual, err := service.New(ctx, testApp).CharacterGetByID(uuid.New())
	assert.Nil(t, actual)
	assert.True(t, ent.IsNotFound(err))
}

func TestCharacterGetByID_WithSoftDeletedCharacter_ReturnsNotFoundError(t *testing.T) {
	refreshDb()

	ctx := context.Background()
	player := seedPlayer(t, ctx)
	now := time.Now()
	character := seedCharacterWithData(t, ctx, player, &ent.Character{Version: 1, DeletedAt: &now})

	actual, err := service.New(ctx, testApp).CharacterGetByID(character.ID)
	assert.Nil(t, actual)
	assert.True(t, ent.IsNotFound(err))
}

func TestCharacterGetByID_WithCharacter_ReturnsCharacter(t *testing.T) {
	refreshDb()

	ctx := context.Background()
	player := seedPlayer(t, ctx)
	character := seedCharacter(t, ctx, player)

	actual, err := service.New(ctx, testApp).CharacterGetByID(character.ID)
	require.NoError(t, err)

	assert.Equal(t, character.ID, actual.ID)
	assert.Equal(t, player.Steamid, actual.Steamid)
	assert.Equal(t, character.Slot, actual.Slot)
	assert.Equal(t, character.Size, actual.Size)
	assert.Equal(t, character.Data, actual.Data)
}

func TestCharacterGetByID_WithCharacter_WithVersions_ReturnsLatestCharacter(t *testing.T) {
	refreshDb()

	ctx := context.Background()
	player := seedPlayer(t, ctx)
	characters := seedCharacters(t, ctx, player, 1, 3)
	require.Len(t, characters, 3)

	actual, err := service.New(ctx, testApp).CharacterGetByID(characters[0].ID)
	require.NoError(t, err)

	character := charById(characters, actual.ID)
	require.NotNil(t, character)

	assert.Equal(t, character.ID, actual.ID)
	assert.Equal(t, player.Steamid, actual.Steamid)
	assert.Equal(t, 1, character.Version)
	assert.Equal(t, character.Slot, actual.Slot)
	assert.Equal(t, character.Size, actual.Size)
	assert.Equal(t, character.Data, actual.Data)
}

func TestCharacterCreate_WithNoPlayer_ReturnsCharacter(t *testing.T) {
	refreshDb()

	ctx := context.Background()
	character := newDeprecatedCharacter()

	actual, err := service.New(ctx, testApp).CharacterCreate(character)
	require.NoError(t, err)

	// Assert returned resource
	assert.NotEqual(t, uuid.Nil, actual.ID)
	assert.Equal(t, character.Steamid, actual.Steamid)
	assert.Equal(t, character.Slot, actual.Slot)
	assert.Equal(t, character.Size, actual.Size)
	assert.Equal(t, character.Data, actual.Data)

	// Assert Player record
	dbPlayer, err := testApp.Client.Player.Query().Where(entPlayer.Steamid(character.Steamid)).Only(ctx)
	require.NoError(t, err)

	assert.NotEqual(t, uuid.Nil, dbPlayer.ID)
	assert.Equal(t, dbPlayer.Steamid, actual.Steamid)

	// Assert Character record
	dbCharacter, err := testApp.Client.Character.Get(ctx, actual.ID)
	require.NoError(t, err)

	assert.Equal(t, dbCharacter.ID, actual.ID)
	assert.Equal(t, dbPlayer.ID, dbCharacter.PlayerID)
	assert.Equal(t, 1, dbCharacter.Version)
	assert.Equal(t, dbCharacter.Slot, actual.Slot)
	assert.Equal(t, dbCharacter.Size, actual.Size)
	assert.Equal(t, dbCharacter.Data, actual.Data)
	assert.Nil(t, dbCharacter.DeletedAt)
}

func TestCharacterCreate_WithPlayer_ReturnsCharacter(t *testing.T) {
	refreshDb()

	ctx := context.Background()
	player := seedPlayer(t, ctx)
	character := newDeprecatedCharacter()
	character.Steamid = player.Steamid

	actual, err := service.New(ctx, testApp).CharacterCreate(character)
	require.NoError(t, err)

	// Assert returned resource
	assert.NotEqual(t, uuid.Nil, actual.ID)
	assert.Equal(t, character.Steamid, actual.Steamid)
	assert.Equal(t, character.Slot, actual.Slot)
	assert.Equal(t, character.Size, actual.Size)
	assert.Equal(t, character.Data, actual.Data)

	// Assert Player record
	dbPlayer, err := testApp.Client.Player.Query().Where(entPlayer.Steamid(character.Steamid)).Only(ctx)
	require.NoError(t, err)

	assert.NotEqual(t, uuid.Nil, dbPlayer.ID)
	assert.Equal(t, dbPlayer.Steamid, actual.Steamid)

	// Assert Character record
	dbCharacter, err := testApp.Client.Character.Get(ctx, actual.ID)
	require.NoError(t, err)

	assert.Equal(t, dbCharacter.ID, actual.ID)
	assert.Equal(t, dbPlayer.ID, dbCharacter.PlayerID)
	assert.Equal(t, 1, dbCharacter.Version)
	assert.Equal(t, dbCharacter.Slot, actual.Slot)
	assert.Equal(t, dbCharacter.Size, actual.Size)
	assert.Equal(t, dbCharacter.Data, actual.Data)
	assert.Nil(t, dbCharacter.DeletedAt)
}

func TestCharacterCreate_WithPlayer_WithSoftDeletedCharacter_ReturnsCharacter(t *testing.T) {
	refreshDb()

	ctx := context.Background()
	player := seedPlayer(t, ctx)
	now := time.Now()
	seedCharacterWithData(t, ctx, player, &ent.Character{Version: 1, DeletedAt: &now})

	character := newDeprecatedCharacter()
	character.Steamid = player.Steamid

	// Assert one soft deleted Character
	count, err := testApp.Client.Character.Query().
		Where(
			entCharacter.And(
				entCharacter.HasPlayerWith(
					entPlayer.Steamid(player.Steamid),
				),
				entCharacter.DeletedAtNotNil(),
			),
		).
		Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	actual, err := service.New(ctx, testApp).CharacterCreate(character)
	require.NoError(t, err)

	// Assert returned resource
	assert.NotEqual(t, uuid.Nil, actual.ID)
	assert.Equal(t, character.Steamid, actual.Steamid)
	assert.Equal(t, character.Slot, actual.Slot)
	assert.Equal(t, character.Size, actual.Size)
	assert.Equal(t, character.Data, actual.Data)

	// Assert Player record
	dbPlayer, err := testApp.Client.Player.Query().Where(entPlayer.Steamid(character.Steamid)).Only(ctx)
	require.NoError(t, err)

	assert.NotEqual(t, uuid.Nil, dbPlayer.ID)
	assert.Equal(t, dbPlayer.Steamid, actual.Steamid)

	// Assert Character record
	dbCharacter, err := testApp.Client.Character.Get(ctx, actual.ID)
	require.NoError(t, err)

	assert.Equal(t, dbCharacter.ID, actual.ID)
	assert.Equal(t, dbPlayer.ID, dbCharacter.PlayerID)
	assert.Equal(t, 1, dbCharacter.Version)
	assert.Equal(t, dbCharacter.Slot, actual.Slot)
	assert.Equal(t, dbCharacter.Size, actual.Size)
	assert.Equal(t, dbCharacter.Data, actual.Data)
	assert.Nil(t, dbCharacter.DeletedAt)

	// Assert only one Character in slot
	count, err = testApp.Client.Character.Query().
		Where(
			entCharacter.HasPlayerWith(
				entPlayer.Steamid(player.Steamid),
			),
		).
		Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestCharacterUpdate_WithNoCharacter_ReturnsNotFoundError(t *testing.T) {
	refreshDb()

	ctx := context.Background()
	character := newDeprecatedCharacter()

	actual, err := service.New(ctx, testApp).CharacterUpdate(uuid.New(), character)
	assert.Nil(t, actual)
	assert.True(t, ent.IsNotFound(err))
}

func TestCharacterUpdate_WithNoBackups_ReturnsUpdatedCharacter(t *testing.T) {
	refreshDb()

	ctx := context.Background()
	player := seedPlayer(t, ctx)
	character := seedCharacter(t, ctx, player)

	updatedCharacter := newDeprecatedCharacter()
	updatedCharacter.Slot = character.Slot
	updatedCharacter.Steamid = player.Steamid

	actual, err := service.New(ctx, testApp).CharacterUpdate(character.ID, updatedCharacter)
	require.NoError(t, err)

	// Assert returned resource
	assert.NotEqual(t, uuid.Nil, actual.ID)
	assert.Equal(t, updatedCharacter.Steamid, actual.Steamid)
	assert.Equal(t, updatedCharacter.Slot, actual.Slot)
	assert.Equal(t, updatedCharacter.Size, actual.Size)
	assert.Equal(t, updatedCharacter.Data, actual.Data)

	// Assert Player record
	dbPlayer, err := testApp.Client.Player.Query().Where(entPlayer.Steamid(updatedCharacter.Steamid)).Only(ctx)
	require.NoError(t, err)

	assert.NotEqual(t, uuid.Nil, dbPlayer.ID)
	assert.Equal(t, dbPlayer.Steamid, actual.Steamid)

	// Assert Character record
	dbCharacter, err := testApp.Client.Character.Get(ctx, actual.ID)
	require.NoError(t, err)

	assert.Equal(t, dbCharacter.ID, actual.ID)
	assert.Equal(t, dbPlayer.ID, dbCharacter.PlayerID)
	assert.Equal(t, 1, dbCharacter.Version)
	assert.Equal(t, dbCharacter.Slot, updatedCharacter.Slot)
	assert.Equal(t, dbCharacter.Size, updatedCharacter.Size)
	assert.Equal(t, dbCharacter.Data, updatedCharacter.Data)
	assert.Nil(t, dbCharacter.DeletedAt)

	// Assert two versions of the Character in slot
	count, err := testApp.Client.Character.Query().
		Where(
			entCharacter.And(
				entCharacter.HasPlayerWith(
					entPlayer.Steamid(player.Steamid),
				),
				entCharacter.Slot(character.Slot),
			),
		).
		Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	// Assert backup Character
	// dbBackup, err := testApp.Client.Character.Query().
	// 	Where(
	// 		entCharacter.And(
	// 			entCharacter.HasPlayerWith(
	// 				entPlayer.Steamid(player.Steamid),
	// 			),
	// 			entCharacter.Version(2),
	// 		),
	// 	).
	// 	Only(ctx)
	// require.NoError(t, err)
	// assert.NotEqual(t, uuid.Nil, dbBackup.ID)
	// assert.NotEqual(t, dbBackup.ID, dbCharacter.ID)
	// assert.Equal(t, dbBackup.PlayerID, dbCharacter.PlayerID)
	// assert.Equal(t, 2, dbBackup.Version)
	// assert.Equal(t, character.Slot, dbBackup.Slot)
	// assert.Equal(t, character.Size, dbBackup.Size)
	// assert.Equal(t, character.Data, dbBackup.Data)
	// assert.Nil(t, dbCharacter.DeletedAt)
}

func TestCharacterUpdate_With15Backups_RemovesExtraBackups_ReturnsUpdatedCharacter(t *testing.T) {
	refreshDb()

	ctx := context.Background()
	player := seedPlayer(t, ctx)
	characters := seedCharacters(t, ctx, player, 1, 15)
	require.Len(t, characters, 15)

	updatedCharacter := newDeprecatedCharacter()
	updatedCharacter.Slot = characters[0].Slot
	updatedCharacter.Steamid = player.Steamid

	actual, err := service.New(ctx, testApp).CharacterUpdate(characters[0].ID, updatedCharacter)
	require.NoError(t, err)

	// Assert returned resource
	assert.NotEqual(t, uuid.Nil, actual.ID)
	assert.Equal(t, updatedCharacter.Steamid, actual.Steamid)
	assert.Equal(t, updatedCharacter.Slot, actual.Slot)
	assert.Equal(t, updatedCharacter.Size, actual.Size)
	assert.Equal(t, updatedCharacter.Data, actual.Data)

	// Assert Player record
	dbPlayer, err := testApp.Client.Player.Query().Where(entPlayer.Steamid(updatedCharacter.Steamid)).Only(ctx)
	require.NoError(t, err)

	assert.NotEqual(t, uuid.Nil, dbPlayer.ID)
	assert.Equal(t, dbPlayer.Steamid, actual.Steamid)

	// Assert Character record
	dbCharacter, err := testApp.Client.Character.Get(ctx, actual.ID)
	require.NoError(t, err)

	assert.Equal(t, dbCharacter.ID, actual.ID)
	assert.Equal(t, dbPlayer.ID, dbCharacter.PlayerID)
	assert.Equal(t, 1, dbCharacter.Version)
	assert.Equal(t, dbCharacter.Slot, updatedCharacter.Slot)
	assert.Equal(t, dbCharacter.Size, updatedCharacter.Size)
	assert.Equal(t, dbCharacter.Data, updatedCharacter.Data)
	assert.Nil(t, dbCharacter.DeletedAt)

	// Assert max 10 versions of the Character in slot
	count, err := testApp.Client.Character.Query().
		Where(
			entCharacter.And(
				entCharacter.HasPlayerWith(
					entPlayer.Steamid(player.Steamid),
				),
				entCharacter.Slot(characters[0].Slot),
			),
		).
		Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 10, count)
}

func TestCharacterDelete_WithNoCharacter_ReturnsNotFoundError(t *testing.T) {
	refreshDb()

	ctx := context.Background()
	err := service.New(ctx, testApp).CharacterDelete(uuid.New())
	assert.True(t, ent.IsNotFound(err))
}

func TestCharacterDelete_WithCharacter_ReturnsNoError(t *testing.T) {
	refreshDb()

	ctx := context.Background()
	player := seedPlayer(t, ctx)
	character := seedCharacter(t, ctx, player)

	err := service.New(ctx, testApp).CharacterDelete(character.ID)
	assert.NoError(t, err)

	dbCharacter, err := testApp.Client.Character.Get(ctx, character.ID)
	require.NoError(t, err)

	assert.NotNil(t, dbCharacter.DeletedAt)
}

func TestCharacterDelete_WithBackups_RemovesAllButLatest_ReturnsNoError(t *testing.T) {
	refreshDb()

	ctx := context.Background()
	player := seedPlayer(t, ctx)
	characters := seedCharacters(t, ctx, player, 1, 3)
	require.Len(t, characters, 3)

	// Assert 3 characters in DB (1 character with 2 backups)
	dbCharacter, err := testApp.Client.Character.Query().
		Where(
			entCharacter.And(
				entCharacter.HasPlayerWith(
					entPlayer.Steamid(player.Steamid),
				),
				entCharacter.Slot(0),
			),
		).
		All(ctx)
	require.NoError(t, err)
	assert.Len(t, dbCharacter, 3)

	err = service.New(ctx, testApp).CharacterDelete(characters[0].ID)
	assert.NoError(t, err)

	dbCharacter, err = testApp.Client.Character.Query().
		Where(
			entCharacter.And(
				entCharacter.HasPlayerWith(
					entPlayer.Steamid(player.Steamid),
				),
				entCharacter.Slot(0),
			),
		).
		All(ctx)
	require.NoError(t, err)

	assert.Len(t, dbCharacter, 1)
	assert.NotNil(t, dbCharacter[0].DeletedAt)
}

func TestCharacterRestore_WithNoCharacter_ReturnsNotFoundError(t *testing.T) {
	refreshDb()

	ctx := context.Background()
	character, err := service.New(ctx, testApp).CharacterRestore(uuid.New())
	assert.Nil(t, character)
	assert.True(t, ent.IsNotFound(err))
}

func TestCharacterRestore_WithCharacter_ReturnsCharacter(t *testing.T) {
	refreshDb()

	ctx := context.Background()
	player := seedPlayer(t, ctx)
	now := time.Now()
	character := seedCharacterWithData(t, ctx, player, &ent.Character{Version: 1, DeletedAt: &now})

	actual, err := service.New(ctx, testApp).CharacterRestore(character.ID)
	require.NoError(t, err)

	assert.Equal(t, character.ID, actual.ID)
	assert.Equal(t, 1, character.Version)
	assert.Equal(t, character.Slot, actual.Slot)
	assert.Equal(t, character.Size, actual.Size)
	assert.Equal(t, character.Data, actual.Data)

	dbCharacter, err := testApp.Client.Character.Get(ctx, character.ID)
	require.NoError(t, err)

	assert.Nil(t, dbCharacter.DeletedAt)
}

func TestCharacterVersions_WithNoCharacter_ReturnsEmptySlice(t *testing.T) {
	refreshDb()

	ctx := context.Background()
	actual, err := service.New(ctx, testApp).CharacterVersions(uuid.NewString(), 0)
	assert.NoError(t, err)
	assert.Empty(t, actual)
}

func TestCharacterVersions_WithCharacter_ReturnsOneCharacter(t *testing.T) {
	refreshDb()

	ctx := context.Background()
	player := seedPlayer(t, ctx)
	character := seedCharacter(t, ctx, player)

	actual, err := service.New(ctx, testApp).CharacterVersions(player.Steamid, 0)
	assert.NoError(t, err)
	require.Len(t, actual, 1)

	assert.Equal(t, character.ID, actual[0].ID)
	assert.Equal(t, character.Slot, actual[0].Slot)
	assert.Equal(t, character.Size, actual[0].Size)
	assert.Equal(t, character.Data, actual[0].Data)
}

func TestCharacterVersions_WithCharacter_ReturnsCharacterAndVersions(t *testing.T) {
	refreshDb()

	ctx := context.Background()
	player := seedPlayer(t, ctx)
	characters := seedCharacters(t, ctx, player, 1, 3)
	require.Len(t, characters, 3)

	actual, err := service.New(ctx, testApp).CharacterVersions(player.Steamid, 0)
	assert.NoError(t, err)
	require.Len(t, actual, 3)
}

func TestCharacterVersions_WithSoftDeletedCharacter_ReturnsOneCharacter(t *testing.T) {
	refreshDb()

	ctx := context.Background()
	player := seedPlayer(t, ctx)
	now := time.Now()
	character := seedCharacterWithData(t, ctx, player, &ent.Character{Version: 1, DeletedAt: &now})

	actual, err := service.New(ctx, testApp).CharacterVersions(player.Steamid, 0)
	assert.NoError(t, err)
	require.Len(t, actual, 1)

	assert.Equal(t, character.ID, actual[0].ID)
	assert.Equal(t, character.Slot, actual[0].Slot)
	assert.Equal(t, character.Size, actual[0].Size)
	assert.Equal(t, character.Data, actual[0].Data)
	assert.NotNil(t, actual[0].DeletedAt)
}

func TestCharacterRollback_WithNoCharacter_ReturnsNotFoundError(t *testing.T) {
	refreshDb()

	ctx := context.Background()
	actual, err := service.New(ctx, testApp).CharacterRollback(uuid.NewString(), 0, 1)
	assert.Nil(t, actual)
	assert.True(t, ent.IsNotFound(err))
}

func TestCharacterRollback_WithInvalidVersion_ReturnsNotFoundError(t *testing.T) {
	refreshDb()

	ctx := context.Background()
	player := seedPlayer(t, ctx)
	seedCharacter(t, ctx, player)

	actual, err := service.New(ctx, testApp).CharacterRollback(player.Steamid, 0, 2)
	assert.Nil(t, actual)
	assert.True(t, ent.IsNotFound(err))
}

func TestCharacterRollback_WhenTargetingCurrentVersion_ReturnsCharacter(t *testing.T) {
	refreshDb()

	ctx := context.Background()
	player := seedPlayer(t, ctx)
	character := seedCharacter(t, ctx, player)

	actual, err := service.New(ctx, testApp).CharacterRollback(player.Steamid, 0, 1)
	assert.NoError(t, err)

	assert.Equal(t, character.ID, actual.ID)
	assert.Equal(t, character.Slot, actual.Slot)
	assert.Equal(t, character.Size, actual.Size)
	assert.Equal(t, character.Data, actual.Data)
}

func TestCharacterRollback_ReturnsCharacter(t *testing.T) {
	refreshDb()

	ctx := context.Background()
	player := seedPlayer(t, ctx)
	characters := seedCharacters(t, ctx, player, 1, 3)
	require.Len(t, characters, 3)

	actual, err := service.New(ctx, testApp).CharacterRollback(player.Steamid, 0, 3)
	assert.NoError(t, err)

	assert.Equal(t, characters[0].ID, actual.ID)
	assert.Equal(t, characters[2].Slot, actual.Slot)
	assert.Equal(t, characters[2].Size, actual.Size)
	assert.Equal(t, characters[2].Data, actual.Data)

	// Assert previous current character is now the latest backup
	dbCharacter, err := testApp.Client.Character.Query().
		Where(
			entCharacter.And(
				entCharacter.HasPlayerWith(
					entPlayer.Steamid(player.Steamid),
				),
				entCharacter.Slot(0),
				entCharacter.Version(4),
			),
		).
		Only(ctx)
	require.NoError(t, err)

	assert.Equal(t, dbCharacter.Slot, characters[0].Slot)
	assert.Equal(t, dbCharacter.Size, characters[0].Size)
	assert.Equal(t, dbCharacter.Data, characters[0].Data)
}

func charById(characters []*ent.Character, id uuid.UUID) *ent.Character {
	for _, c := range characters {
		if c.ID == id {
			return c
		}
	}
	return nil
}

func seedPlayer(t *testing.T, ctx context.Context) *ent.Player {
	player, err := testApp.Client.Player.Create().
		SetSteamid(uuid.NewString()).
		Save(ctx)
	require.NoError(t, err, "failed to create Player")

	return player
}

func seedCharacter(t *testing.T, ctx context.Context, player *ent.Player) *ent.Character {
	if player == nil {
		panic("failed to create Character; Player is nil")
	}

	character := &ent.Character{
		Version: 1,
		Slot:    0,
		Size:    rand.Intn(4096),
		Data:    "data",
	}

	return seedCharacterWithData(t, ctx, player, character)
}

func seedCharacters(t *testing.T, ctx context.Context, player *ent.Player, cCount, vCount int) []*ent.Character {
	if player == nil {
		panic("failed to create Character; Player is nil")
	}

	characters := make([]*ent.Character, 0, cCount*vCount)
	for slot := 0; slot < cCount; slot++ {
		for version := 1; version <= vCount; version++ {
			characters = append(characters, seedCharacterWithData(t, ctx, player, &ent.Character{
				Slot:    slot,
				Version: version,
				Size:    rand.Intn(4096),
				Data:    "data",
			}))
		}
	}

	return characters
}

func seedCharacterWithData(t *testing.T, ctx context.Context, player *ent.Player, c *ent.Character) *ent.Character {
	if player == nil || c == nil {
		panic("failed to create Character; Player or character is nil")
	}

	builder := testApp.Client.Character.Create().
		SetPlayer(player).
		SetVersion(c.Version).
		SetSize(c.Size).
		SetSlot(c.Slot).
		SetData(c.Data)

	if c.DeletedAt != nil {
		builder = builder.SetNillableDeletedAt(c.DeletedAt)
	}

	character, err := builder.Save(ctx)
	require.NoError(t, err, "failed to create Character")

	return character
}

func newDeprecatedCharacter() ent.DeprecatedCharacter {
	return ent.DeprecatedCharacter{
		Steamid: uuid.NewString(),
		Slot:    0,
		Size:    rand.Intn(4096),
		Data:    "data",
	}
}

// func (s *service) CharacterCreate(newChar ent.DeprecatedCharacter) (*ent.DeprecatedCharacter, error) {
// 	var char *ent.Character
// 	err := txn(s.ctx, s.client, func(tx *ent.Tx) error {
// 		player, err := s.client.Player.Query().
// 			Where(
// 				player.Steamid(newChar.Steamid),
// 			).
// 			Only(s.ctx)
// 		if err != nil {
// 			if !ent.IsNotFound(err) {
// 				return err
// 			}

// 			// Create Player if one doesn't exist
// 			player, err = s.client.Player.Create().
// 				SetSteamid(newChar.Steamid).
// 				Save(s.ctx)
// 			if err != nil {
// 				return err
// 			}
// 		}

// 		// Hard delete characters taking the requested slot
// 		found, err := s.client.Character.Query().
// 			Where(
// 				character.And(
// 					character.PlayerID(player.ID),
// 					character.Slot(newChar.Slot),
// 				),
// 			).
// 			Exist(s.ctx)
// 		if err != nil {
// 			return err
// 		}

// 		if found {
// 			_, err = s.client.Character.Delete().
// 				Where(
// 					character.And(
// 						character.PlayerID(player.ID),
// 						character.Slot(newChar.Slot),
// 					),
// 				).
// 				Exec(s.ctx)
// 			if err != nil {
// 				return err
// 			}
// 		}

// 		// Create new character
// 		c, err := s.client.Character.Create().
// 			SetPlayer(player).
// 			SetSlot(newChar.Slot).
// 			SetSize(newChar.Size).
// 			SetData(newChar.Data).
// 			SetVersion(1).
// 			Save(s.ctx)
// 		if err != nil {
// 			return err
// 		}

// 		char = c
// 		return nil
// 	})
// 	if err != nil {
// 		return nil, err
// 	}

// 	return charToDepChar(newChar.Steamid, char), nil
// }

// func (s *service) CharacterUpdate(uid uuid.UUID, updateChar ent.DeprecatedCharacter) (*ent.DeprecatedCharacter, error) {
// 	var char *ent.Character
// 	err := txn(s.ctx, s.client, func(tx *ent.Tx) error {
// 		// Get the current character
// 		current, err := s.client.Character.Get(s.ctx, uid)
// 		if err != nil {
// 			return err
// 		}

// 		// Get the latest backup version
// 		latest, err := s.client.Character.Query().
// 			Select(character.FieldVersion).
// 			Where(character.ID(uid)).
// 			Order(ent.Desc(character.FieldVersion)).
// 			First(s.ctx)
// 		if err != nil {
// 			return err
// 		}

// 		// Backup the current version
// 		_, err = s.client.Character.Create().
// 			SetPlayerID(current.PlayerID).
// 			SetVersion(latest.Version + 1).
// 			SetSlot(current.Slot).
// 			SetSize(current.Size).
// 			SetData(current.Data).
// 			Save(s.ctx)
// 		if err != nil {
// 			return err
// 		}

// 		// Update the character
// 		c, err := s.client.Character.UpdateOneID(uid).
// 			SetSize(updateChar.Size).
// 			SetData(updateChar.Data).
// 			Save(s.ctx)
// 		if err != nil {
// 			return err
// 		}

// 		fmt.Printf("%v", c)

// 		// Get all backup characters
// 		all, err := s.client.Character.Query().
// 			Where(
// 				character.And(
// 					character.PlayerID(c.PlayerID),
// 					character.Slot(c.Slot),
// 					character.VersionNEQ(c.Version),
// 				),
// 			).
// 			Order(ent.Desc(character.FieldCreatedAt)).
// 			All(s.ctx)
// 		if err != nil {
// 			return err
// 		}

// 		// Delete all characters beyond 10 backups
// 		if len(all) > 10 {
// 			for _, old := range all[10:] {
// 				if err := s.client.Character.DeleteOneID(old.ID).Exec(s.ctx); err != nil {
// 					return err
// 				}
// 			}
// 		}

// 		char = c
// 		return nil
// 	})
// 	if err != nil {
// 		return nil, err
// 	}

// 	return charToDepChar(updateChar.Steamid, char), nil
// }

// func (s *service) CharacterDelete(uid uuid.UUID) error {
// 	return txn(s.ctx, s.client, func(tx *ent.Tx) error {
// 		// Get Current version
// 		char, err := s.client.Character.Get(s.ctx, uid)
// 		if err != nil {
// 			return err
// 		}

// 		// Soft delete
// 		_, err = char.Update().
// 			SetDeletedAt(time.Now()).
// 			Save(s.ctx)
// 		if err != nil {
// 			return err
// 		}

// 		// Hard delete backups
// 		_, err = s.client.Character.Delete().
// 			Where(
// 				character.And(
// 					character.PlayerID(char.PlayerID),
// 					character.Slot(char.Slot),
// 					character.VersionNEQ(char.Version),
// 				),
// 			).
// 			Exec(s.ctx)
// 		if err != nil {
// 			return err
// 		}

// 		return nil
// 	})
// }

// func (s *service) CharacterRestore(uid uuid.UUID) (*ent.DeprecatedCharacter, error) {
// 	char, err := s.client.Character.UpdateOneID(uid).
// 		ClearDeletedAt().
// 		Save(s.ctx)
// 	if err != nil {
// 		return nil, err
// 	}

// 	player, err := char.QueryPlayer().Only(s.ctx)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return charToDepChar(player.Steamid, char), nil
// }

// func (s *service) CharacterVersions(sid string, slot int) ([]*ent.Character, error) {
// 	chars, err := s.client.Character.Query().
// 		Where(
// 			character.And(
// 				character.HasPlayerWith(player.Steamid(sid)),
// 				character.Slot(slot),
// 			),
// 		).
// 		All(s.ctx)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return chars, nil
// }

// func (s *service) CharacterRollback(sid string, slot, version int) (*ent.DeprecatedCharacter, error) {
// 	var char *ent.Character
// 	err := txn(s.ctx, s.client, func(tx *ent.Tx) error {
// 		// Get the current character
// 		targeted, err := s.client.Character.Query().
// 			Where(
// 				character.And(
// 					character.HasPlayerWith(player.Steamid(sid)),
// 					character.Slot(slot),
// 					character.Version(version),
// 				),
// 			).
// 			First(s.ctx)
// 		if err != nil {
// 			return err
// 		}

// 		// Get the current character
// 		current, err := s.client.Character.Query().
// 			Where(
// 				character.And(
// 					character.HasPlayerWith(player.Steamid(sid)),
// 					character.Version(1),
// 				),
// 			).
// 			First(s.ctx)
// 		if err != nil {
// 			return err
// 		}

// 		// Get the latest backup version
// 		latest, err := s.client.Character.Query().
// 			Select(character.FieldVersion).
// 			Where(character.HasPlayerWith(player.Steamid(sid))).
// 			Order(ent.Desc(character.FieldVersion)).
// 			First(s.ctx)
// 		if err != nil {
// 			return err
// 		}

// 		// Backup the current version
// 		_, err = s.client.Character.Create().
// 			SetPlayerID(current.PlayerID).
// 			SetVersion(latest.Version + 1).
// 			SetSlot(current.Slot).
// 			SetSize(current.Size).
// 			SetData(current.Data).
// 			Save(s.ctx)
// 		if err != nil {
// 			return err
// 		}

// 		// Update the character
// 		c, err := current.Update().
// 			SetSize(targeted.Size).
// 			SetData(targeted.Data).
// 			Save(s.ctx)
// 		if err != nil {
// 			return err
// 		}

// 		char = c
// 		return nil
// 	})
// 	if err != nil {
// 		return nil, err
// 	}

// 	return charToDepChar(sid, char), nil
// }

// func charToDepChar(s string, c *ent.Character) *ent.DeprecatedCharacter {
// 	return &ent.DeprecatedCharacter{
// 		ID:      c.ID,
// 		Steamid: s,
// 		Slot:    c.Slot,
// 		Size:    c.Size,
// 		Data:    c.Data,
// 	}
// }

// func charsToDepChars(c []*ent.Character) []*ent.DeprecatedCharacter {
// 	deps := make([]*ent.DeprecatedCharacter, len(c))
// 	for i := range c {
// 		deps[i] = charToDepChar(c[i].Edges.Player.Steamid, c[i])
// 	}
// 	return deps
// }
