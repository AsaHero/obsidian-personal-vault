package compute

type Query struct {
	command   CommandID
	arguments []string
}

func NewQuery(cmd CommandID, args ...string) Query {
	return Query{
		command:   cmd,
		arguments: args,
	}
}

func (q *Query) CommandID() CommandID {
	return q.command
}

func (q *Query) Arguments() []string {
	return q.arguments
}
