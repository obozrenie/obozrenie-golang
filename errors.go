package main

import "errors"

// Errors contains error exceptions that Obozrenie may throw.
var errWrongPassword = errors.New("Wrong password")
var errNoGamesSpecified = errors.New("No games specified")
var errNoProxy = errors.New("Specified proxy does not exist")
var errEmptyProxyData = errors.New("Empty proxy data")
var errNoAdapter = errors.New("Specified adapter does not exist")
var errInvalidID = errors.New("Invalid ID")
var errInvalidGameID = errors.New("Please specify a valid game id")
var errGameExists = errors.New("Specified game already exists in the database")
var errNoSuchGame = errors.New("Specified game is not found in the database")
var errUnknownGameID = errors.New("Specified game ID is not found in the database")
var errinvalidIDList = errors.New("Please specify a list of valid game IDs")
var errMalformedEntry = errors.New("Malformed server entry")
