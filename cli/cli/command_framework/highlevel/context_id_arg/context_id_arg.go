package context_id_arg

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/uuid_generator"
	"github.com/kurtosis-tech/kurtosis/contexts-config-store/api/golang"
	"github.com/kurtosis-tech/kurtosis/contexts-config-store/api/golang/generated"
	"github.com/kurtosis-tech/kurtosis/contexts-config-store/store"
	"github.com/kurtosis-tech/stacktrace"
	"sort"
	"strings"
)

const (
	uuidOrNameDelimiter = ", "

	defaultIsRequired = false
	defaultValueEmpty = ""
)

// NewContextIdentifierArg pre-builds context identifier arg which has tab-completion and validation ready out-of-the-box
func NewContextIdentifierArg(
	// The arg key where this context identifier argument will be stored
	argKey string,
	isGreedy bool,
) *args.ArgConfig {

	validate := getValidationFunc(argKey, isGreedy)

	return &args.ArgConfig{
		Key:                   argKey,
		IsOptional:            defaultIsRequired,
		DefaultValue:          defaultValueEmpty,
		IsGreedy:              isGreedy,
		ValidationFunc:        validate,
		ArgCompletionProvider: args.NewManualCompletionsProvider(getCompletions),
	}
}

// Make best-effort attempt to get context names
func getCompletions(ctx context.Context, flags *flags.ParsedFlags, previousArgs *args.ParsedArgs) ([]string, error) {
	contextsConfigStore := store.GetContextsConfigStore()

	kurtosisContextsConfig, err := contextsConfigStore.GetKurtosisContextsConfig()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting Kurtosis contexts configuration")
	}

	var storedContextNames []string
	for _, storedContextInfo := range kurtosisContextsConfig.GetContexts() {
		storedContextNames = append(storedContextNames, storedContextInfo.GetName())
	}

	// we sort them individually
	sort.Strings(storedContextNames)
	return storedContextNames, nil
}

// Context identifier validation function
func getValidationFunc(argKey string, isGreedy bool) func(context.Context, *flags.ParsedFlags, *args.ParsedArgs) error {
	return func(ctx context.Context, flags *flags.ParsedFlags, args *args.ParsedArgs) error {

		var contextIdentifiersToValidate []string
		if isGreedy {
			contextIdentifiers, err := args.GetGreedyArg(argKey)
			if err != nil {
				return stacktrace.Propagate(err, "Expected a value for greedy arg '%v' but didn't find any", argKey)
			}
			contextIdentifiersToValidate = append(contextIdentifiersToValidate, contextIdentifiers...)
		} else {
			contextIdentifier, err := args.GetNonGreedyArg(argKey)
			if err != nil {
				return stacktrace.Propagate(err, "Expected a value for non-greedy arg '%v' but didn't find one", argKey)
			}
			contextIdentifiersToValidate = append(contextIdentifiersToValidate, contextIdentifier)
		}

		_, err := GetContextUuidForContextIdentifier(contextIdentifiersToValidate)
		if err != nil {
			return stacktrace.Propagate(err, "Error finding context matching the provided identifiers")
		}
		return nil
	}
}

func GetContextUuidForContextIdentifier(contextIdentifiers []string) (map[string]*generated.ContextUuid, error) {
	contextsConfigStore := store.GetContextsConfigStore()
	kurtosisContextsConfig, err := contextsConfigStore.GetKurtosisContextsConfig()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting Kurtosis contexts configuration")
	}

	// Index all context identifiers in the following three maps
	contextsByUuid := map[string]bool{}              // set of contexts UUID currently stored
	contextsByShortenedUuid := map[string][]string{} // mapping of Shortened UUID -> UUID for stored contexts
	contextsByName := map[string][]string{}          // mapping of Name -> UUID for stored contexts
	for _, storedContextInfo := range kurtosisContextsConfig.GetContexts() {
		storedContextUuid := storedContextInfo.GetUuid().GetValue()
		storedContextShortenedUuid := uuid_generator.ShortenedUUIDString(storedContextUuid)
		storedContextName := storedContextInfo.GetName()

		contextsByUuid[storedContextName] = true
		contextsByShortenedUuid[storedContextShortenedUuid] = append(contextsByShortenedUuid[storedContextShortenedUuid], storedContextUuid)
		contextsByName[storedContextName] = append(contextsByName[storedContextName], storedContextUuid)
	}

	contextUuids := map[string]*generated.ContextUuid{}
	for _, contextIdentifier := range contextIdentifiers {
		if contextsWithMatchingNames, found := contextsByName[contextIdentifier]; found {
			if len(contextsWithMatchingNames) > 1 {
				return nil, stacktrace.NewError("Found multiple contexts matching context name: '%s': '%s'. This is ambiguous", contextIdentifier, strings.Join(contextsWithMatchingNames, uuidOrNameDelimiter))
			}
			contextUuids[contextIdentifier] = golang.NewContextUuid(contextsWithMatchingNames[0])
			continue
		}
		// check if full UUID is a match
		if _, found := contextsByUuid[contextIdentifier]; found {
			contextUuids[contextIdentifier] = golang.NewContextUuid(contextIdentifier)
			continue
		}
		// check if shortened UUID is a match
		if contextsWithMatchingShortenedUuids, found := contextsByShortenedUuid[contextIdentifier]; found {
			if len(contextsWithMatchingShortenedUuids) > 1 {
				return nil, stacktrace.NewError("Found multiple contexts matching shortened UUID: '%s': '%s'. THis is ambiguous", contextIdentifier, strings.Join(contextsWithMatchingShortenedUuids, uuidOrNameDelimiter))
			}
			contextUuids[contextIdentifier] = golang.NewContextUuid(contextsWithMatchingShortenedUuids[0])
			continue
		}
		return nil, stacktrace.NewError("No context found for identifier '%v'", contextIdentifier)
	}
	return contextUuids, nil
}
