import { Component, OnInit, Input } from '@angular/core';

@Component({
  selector: 'app-home-navigation',
  templateUrl: './home-navigation.component.html',
  styleUrls: ['./home-navigation.component.scss']
})
export class HomeNavigationComponent implements OnInit {

  constructor() { }

  @Input() section: string;

  ngOnInit() {
  }

}
