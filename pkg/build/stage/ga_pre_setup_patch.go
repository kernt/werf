package stage

func NewGAPreSetupPatchStage(baseStageOptions *NewBaseStageOptions) *GAPreSetupPatchStage {
	s := &GAPreSetupPatchStage{}
	s.GARelatedStage = newGARelatedStage(baseStageOptions)
	return s
}

type GAPreSetupPatchStage struct {
	*GARelatedStage
}

func (s *GAPreSetupPatchStage) Name() StageName {
	return GAPreSetupPatch
}

func (s *GAPreSetupPatchStage) GetRelatedStageName() StageName {
	return Setup
}