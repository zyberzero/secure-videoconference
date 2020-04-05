using System.Threading.Tasks;
using meetingadmin.Domain;
using meetingadmin.Services.Meetings;
using Microsoft.AspNetCore.Mvc;
using Microsoft.Extensions.Logging;

namespace meetingadmin.Controllers
{
    [ApiController]
    [Route("[controller]")]
    public class MeetingController : ControllerBase
    {
        private readonly ILogger<MeetingController> logger;
        private readonly IMeetingService meetingService;

        public MeetingController(ILogger<MeetingController> logger, IMeetingService meetingService)
        {
            this.logger = logger;
            this.meetingService = meetingService;
        }

        [HttpPost]
        public async Task<IActionResult> CreateMeeting(Meeting meeting)
        {
            var meetingRoom = await this.meetingService.AddMeeting(meeting);
            return Ok(meetingRoom);
        }
    }
}
