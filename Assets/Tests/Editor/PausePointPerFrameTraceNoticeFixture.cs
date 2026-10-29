namespace io.github.hatayama.UnityCliLoop.Tests.PausePointToolsFixtures
{
    internal sealed class PerFrameTraceNoticeFixture
    {
        private int _probe;

        public void Update()
        {
            // per-frame-trace-notice-probe-unique
            _probe = 1;
        }

        // Reads _probe so the compiler does not report it as assigned but never used (CS0414).
        public int Probe => _probe;
    }
}
