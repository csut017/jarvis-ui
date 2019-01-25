import { Component, OnInit } from '@angular/core';
import { SpeechService } from '../services/speech.service';
import { TimeService } from '../services/time.service';

@Component({
  selector: 'app-home',
  templateUrl: './home.component.html',
  styleUrls: ['./home.component.scss']
})
export class HomeComponent implements OnInit {

  constructor(private speech: SpeechService,
    private time: TimeService) { }

  ngOnInit() {
  }

  sayWelcome() {
    const dayPart = this.time.getPartOfDay(),
      timePart = this.time.getFriendlyTime();
    this.speech.say('Good ' + dayPart + ' Craig, it is ' + timePart);
  }
}
