using System;
using System.Runtime.CompilerServices;

namespace io.github.hatayama.UnityCliLoop.Tests.Editor.HotReload
{
    /// <summary>
    /// Other file of <see cref="HotReloadAddedMemberPartialHost"/> so the compiled type
    /// has members this file does not declare.
    /// </summary>
    public partial class HotReloadAddedMemberPartialHost
    {
        public int PartialOtherProperty { get; set; }

#pragma warning disable CS0067
        public event Action PartialOtherEvent;
#pragma warning restore CS0067

        [MethodImpl(MethodImplOptions.NoInlining)]
        public int PartialOtherFile()
        {
            return 3;
        }
    }
}
