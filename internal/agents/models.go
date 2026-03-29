package agents

import "time"

type agentModel struct {
	ID           string    `json:"id"`
	Role         Role      `json:"role"`
	Address      string    `json:"address"`
	Capabilities []string  `json:"capabilities"`
	RegisteredAt time.Time `json:"registered_at"`
}

type tokenModel struct {
	AgentID   string    `json:"agent_id"`
	CreatedAt time.Time `json:"created_at"`
}

func (a *agentModel) toDomain() *Agent {
	if a == nil {
		return nil
	}

	return &Agent{
		AgentInput: AgentInput{
			ID:           a.ID,
			Role:         a.Role,
			Address:      a.Address,
			Capabilities: a.Capabilities,
		},
		RegisteredAt: a.RegisteredAt,
	}
}
