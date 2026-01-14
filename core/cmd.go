package core

type Command struct {
	Command string
	Args    []string
}

func GetCommand(tokens []string) Command {
	return Command{
		Command: tokens[0],
		Args:    tokens[1:],
	}
}
