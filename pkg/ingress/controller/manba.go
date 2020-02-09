package controller

// OnUpdate is called periodically by syncQueue to keep the configuration in sync.
// // returning nil implies the synchronization finished correctly.
// // Returning an error means requeue the update.
func (m *ManbaController) OnUpdate() error {
	return nil
}
